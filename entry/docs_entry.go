// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	rkembed "github.com/rookie-ninja/rk-entry"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	specFiles        = make(map[string]string, 0)
	specFileContents = ``
)

// Inner struct used while initializing swagger entry.
type docsConfig struct {
	Specs []*spec `json:"specs" yaml:"specs"`
	Style struct {
		Theme       string `yaml:"theme" json:"theme"`
		RenderStyle string `yaml:"renderStyle" json:"renderStyle"`
		AllowTry    bool   `yaml:"allowTry" json:"allowTry"`
		BgColor     string `yaml:"bgColor" json:"bgColor"`
	} `json:"style" yaml:"style"`
}

// Inner struct used while initializing open API entry.
type spec struct {
	Name string `json:"name" yaml:"name"`
	Url  string `json:"url" yaml:"url"`
}

// BootDocs Bootstrap config of swagger.
// 1: Enabled: Enable swagger.
// 2: Path: Swagger path accessible from restful API.
// 3: SpecPath: The path of where swagger or open API spec file was located.
// 4: Headers: The headers that would be added into each API response.
type BootDocs struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	Path     string   `yaml:"path" json:"path"`
	SpecPath string   `yaml:"specPath" json:"specPath"`
	Headers  []string `yaml:"headers" json:"headers"`
	Style    struct {
		Theme string `yaml:"theme" json:"theme"`
	} `yaml:"style" json:"style"`
	Debug struct {
		Enabled bool   `yaml:"enabled" json:"enabled"`
		Path    string `yaml:"path" json:"path"`
	} `yaml:"debug" json:"debug"`
}

// DocsEntry implements rkentry.Entry interface.
type DocsEntry struct {
	entryName        string            `json:"-" yaml:"-"`
	entryType        string            `json:"-" yaml:"-"`
	entryDescription string            `json:"-" yaml:"-"`
	SpecPath         string            `json:"-" yaml:"-"`
	Path             string            `json:"-" yaml:"-"`
	Headers          map[string]string `json:"-" yaml:"-"`
	Debug            struct {
		Enabled bool `yaml:"-" json:"-"`
	} `yaml:"-" json:"-"`
	Style struct {
		Theme string `yaml:"-" json:"-"`
	} `yaml:"-" json:"-"`
	embedFS *embed.FS `json:"-" yaml:"-"`
}

func WithNameDocsEntry(name string) DocsEntryOption {
	return func(entry *DocsEntry) {
		entry.entryName = name
	}
}

func RegisterDocsEntry(boot *BootDocs, opts ...DocsEntryOption) *DocsEntry {
	var docsEntry *DocsEntry
	if boot.Enabled {
		// Init swagger custom headers from config
		headers := make(map[string]string, 0)
		for i := range boot.Headers {
			header := boot.Headers[i]
			tokens := strings.Split(header, ":")
			if len(tokens) == 2 {
				headers[tokens[0]] = tokens[1]
			}
		}

		docsEntry = &DocsEntry{
			entryName:        "DocsEntry",
			entryType:        "DocsEntry",
			entryDescription: "Internal RK entry for documentation UI.",
			Path:             boot.Path,
			SpecPath:         boot.SpecPath,
			Headers:          headers,
		}

		for i := range opts {
			opts[i](docsEntry)
		}

		if len(docsEntry.Path) < 1 {
			docsEntry.Path = "/docs"
		}

		docsEntry.Debug.Enabled = boot.Debug.Enabled

		docsEntry.Style.Theme = strings.ToLower(boot.Style.Theme)
		if docsEntry.Style.Theme != "light" && docsEntry.Style.Theme != "dark" {
			docsEntry.Style.Theme = "light"
		}

		// Deal with Path
		// add "/" at start and end side if missing
		docsEntry.Path = slashPath(docsEntry.Path)
	}

	return docsEntry
}

type DocsEntryOption func(entry *DocsEntry)

func (entry *DocsEntry) Bootstrap(ctx context.Context) {
	// init swagger configs
	entry.initDocsConfig()
}

func (entry *DocsEntry) Interrupt(ctx context.Context) {}

func (entry *DocsEntry) GetName() string {
	return entry.entryName
}

func (entry *DocsEntry) GetType() string {
	return entry.entryType
}

func (entry *DocsEntry) GetDescription() string {
	return entry.entryDescription
}

func (entry *DocsEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry
func (entry *DocsEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.GetName(),
		"type":        entry.GetType(),
		"description": entry.GetDescription(),
		"specPath":    entry.SpecPath,
		"path":        entry.Path,
		"Headers":     entry.Headers,
		"debug":       entry.Debug.Enabled,
	}

	return json.Marshal(m)
}

// UnmarshalJSON Unmarshal entry
func (entry *DocsEntry) UnmarshalJSON([]byte) error {
	return nil
}

func (entry *DocsEntry) SetEmbedFS(fs *embed.FS) {
	entry.embedFS = fs
}

// ConfigFileHandler handler for swagger config files.
func (entry *DocsEntry) ConfigFileHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		p := strings.TrimPrefix(strings.TrimSuffix(request.URL.Path, "/"), strings.TrimSuffix(entry.Path, "/"))
		p = strings.TrimSuffix(p, "/")
		p = strings.TrimPrefix(p, "/")

		writer.Header().Set("cache-control", "no-cache")

		for k, v := range entry.Headers {
			writer.Header().Set(k, v)
		}

		switch p {
		case "":
			if file := readFile("assets/docs/index.html", &rkembed.AssetsFS, false); len(file) < 1 {
				http.Error(writer, "Internal server error", http.StatusInternalServerError)
			} else {
				http.ServeContent(writer, request, "index.html", time.Now(), bytes.NewReader(file))
			}
		case "specs":
			http.ServeContent(writer, request, "specs", time.Now(), strings.NewReader(specFileContents))
		default:
			value, ok := swaggerJsonFiles[p]
			if ok {
				http.ServeContent(writer, request, p, time.Now(), strings.NewReader(value))
				return
			}

			http.NotFound(writer, request)
		}
	}
}

// Init swagger or open API spec config.
// This function do the things bellow:
// 1: List files from entry.SpecPath.
// 2: Read user swagger json files and deduplicate.
// 3: Assign swagger or open API spec contents into specFileContents variable
func (entry *DocsEntry) initDocsConfig() {
	config := &docsConfig{
		Specs: []*spec{},
	}

	if len(entry.SpecPath) > 0 {
		// 1: Add user API swagger JSON
		entry.listFilesWithSuffix(config, entry.SpecPath, false)
	} else {
		// try to read from default directories
		// - docs
		// - api/gen/v1
		// - api/gen
		entry.listFilesWithSuffix(config, "docs", true)
		entry.listFilesWithSuffix(config, "api/gen/v1", true)
		entry.listFilesWithSuffix(config, "api/gen", true)
	}

	// 2: Add rk common APIs
	if len(swAssetsFile) > 0 {
		key := entry.entryName + "-rk-common.swagger.json"
		// add common service json file
		specFiles[key] = string(swAssetsFile)
		config.Specs = append(config.Specs, &spec{
			Name: key,
			Url:  path.Join(entry.Path, key),
		})
	}

	// 3: Assign style
	config.Style.Theme = entry.Style.Theme
	config.Style.RenderStyle = "focused"
	config.Style.AllowTry = false
	if config.Style.Theme == "light" {
		config.Style.BgColor = "#FAFAFA"
	}

	if entry.Debug.Enabled {
		config.Style.RenderStyle = "focused"
		config.Style.AllowTry = true
	}

	// 4: Marshal to swagger-config.json
	bytes, err := json.Marshal(config)
	if err != nil {
		ShutdownWithError(err)
	}

	specFileContents = string(bytes)
}

// List files with .json suffix and store them into swaggerJsonFiles variable.
func (entry *DocsEntry) listFilesWithSuffix(config *docsConfig, specPath string, ignoreError bool) {
	suffix := ".json"

	if entry.embedFS != nil {
		// 1: read dir
		files, err := entry.embedFS.ReadDir(specPath)
		if err != nil && !ignoreError {
			return
		}

		for i := range files {
			file := files[i]
			if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
				bytes, err := entry.embedFS.ReadFile(path.Join(specPath, file.Name()))
				key := entry.entryName + "-" + file.Name()

				if err != nil && !ignoreError {
					ShutdownWithError(err)
				}

				swaggerJsonFiles[key] = string(bytes)

				config.Specs = append(config.Specs, &spec{
					Name: key,
					Url:  path.Join(entry.Path, key),
				})
			}
		}

		return
	}

	// re-path it with working directory if not absolute path
	if !path.IsAbs(specPath) {
		wd, _ := os.Getwd()
		specPath = path.Join(wd, specPath)
	}

	fmt.Println(specPath)

	files, err := ioutil.ReadDir(specPath)
	if err != nil && !ignoreError {
		return
	}

	for i := range files {
		file := files[i]
		if !file.IsDir() && strings.HasSuffix(file.Name(), suffix) {
			bytes, err := ioutil.ReadFile(path.Join(specPath, file.Name()))
			key := entry.entryName + "-" + file.Name()

			if err != nil && !ignoreError {
				ShutdownWithError(err)
			}

			swaggerJsonFiles[key] = string(bytes)

			config.Specs = append(config.Specs, &spec{
				Name: key,
				Url:  path.Join(entry.Path, key),
			})
		}
	}
}