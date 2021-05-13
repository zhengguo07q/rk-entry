// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"encoding/json"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-query"
)

const (
	AppNameDefault          = "rkApp"
	VersionDefault          = "v0.0.0"
	LangDefault             = "golang"
	DescriptionDefault      = "rk application"
	AppInfoEntryName        = "AppInfoDefault"
	AppInfoEntryType        = "AppInfoEntry"
	AppInfoEntryDescription = "Internal RK entry which describes application with fields of appName, version and etc."
)

// Bootstrap config of application's basic information.
// 1: AppName: Application name which refers to go process.
// 2: Version: Application version.
// 3: Description: Description of application itself.
// 5: Keywords: A set of words describe application.
// 6: HomeUrl: Home page URL.
// 7: IconUrl: Application Icon URL.
// 8: Maintainers: Maintainers of application.
// 9: DocsUrl: A set of URLs of documentations of application.
type BootConfigAppInfo struct {
	RK struct {
		AppName     string   `yaml:"appName" json:"appName"`
		Version     string   `yaml:"version" json:"version"`
		Description string   `yaml:"description" json:"description"`
		Keywords    []string `yaml:"keywords" json:"keywords"`
		HomeUrl     string   `yaml:"homeUrl" json:"homeUrl"`
		IconUrl     string   `yaml:"iconUrl" json:"iconUrl"`
		DocsUrl     []string `yaml:"docsUrl" json:"docsUrl"`
		Maintainers []string `yaml:"maintainers" json:"maintainers"`
	} `yaml:"rk"`
}

// AppInfo Entry contains bellow fields.
// 1: AppName: Application name which refers to go process
// 2: Version: Application version
// 3: Lang: Programming language <NOT configurable!>
// 4: Description: Description of application itself
// 5: Keywords: A set of words describe application
// 6: HomeUrl: Home page URL
// 7: IconUrl: Application Icon URL
// 8: DocsUrl: A set of URLs of documentations of application
// 9: Maintainers: Maintainers of application
type AppInfoEntry struct {
	EntryName        string   `json:"entryName" yaml:"entryName"`
	EntryType        string   `json:"entryType" yaml:"entryType"`
	EntryDescription string   `json:"entryDescription" yaml:"entryDescription"`
	Description      string   `json:"description" yaml:"description"`
	AppName          string   `json:"appName" yaml:"appName"`
	Version          string   `json:"version" yaml:"version"`
	Lang             string   `json:"lang" yaml:"lang"`
	Keywords         []string `json:"keywords" yaml:"keywords"`
	HomeUrl          string   `json:"homeUrl" yaml:"homeUrl"`
	IconUrl          string   `json:"iconUrl" yaml:"iconUrl"`
	DocsUrl          []string `json:"docsUrl" yaml:"docsUrl"`
	Maintainers      []string `json:"maintainers" yaml:"maintainers"`
}

// Generate a AppInfo entry with default fields.
// 1: AppName: rkApp
// 2: Version: v0.0.0
// 4: Description: rk application
// 3: Lang: golang
// 5: Keywords: []
// 6: HomeURL: ""
// 7: IconURL: ""
// 8: Maintainers: []
// 9: DocsURL: []]
func AppInfoEntryDefault() *AppInfoEntry {
	return &AppInfoEntry{
		EntryName:        AppInfoEntryName,
		EntryType:        AppInfoEntryType,
		EntryDescription: AppInfoEntryDescription,
		AppName:          AppNameDefault,
		Version:          VersionDefault,
		Lang:             LangDefault,
		Description:      DescriptionDefault,
		Keywords:         []string{},
		HomeUrl:          "",
		IconUrl:          "",
		DocsUrl:          []string{},
		Maintainers:      []string{},
	}
}

// AppInfo Entry Option which used while registering entry from codes.
type AppInfoEntryOption func(*AppInfoEntry)

// Provide application name.
func WithAppNameAppInfo(AppName string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.AppName = AppName
	}
}

// Provide version.
func WithVersionAppInfo(version string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Version = version
	}
}

// Provide description.
func WithDescriptionAppInfo(description string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Description = description
	}
}

// Provide home page URL.
func WithHomeUrlAppInfo(homeUrl string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.HomeUrl = homeUrl
	}
}

// Provide icon URL.
func WithIconUrlAppInfo(iconUrl string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.IconUrl = iconUrl
	}
}

// Provide keywords.
func WithKeywordsAppInfo(keywords ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Keywords = append(entry.Keywords, keywords...)
	}
}

// Provide documentation URLs.
func WithDocsUrlAppInfo(docsURL ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.DocsUrl = append(entry.DocsUrl, docsURL...)
	}
}

// Provide maintainers.
func WithMaintainersAppInfo(maintainers ...string) AppInfoEntryOption {
	return func(entry *AppInfoEntry) {
		entry.Maintainers = append(entry.Maintainers, maintainers...)
	}
}

// Implements rkentry.EntryRegFunc which generate RKEntry based on boot configuration file.
func RegisterAppInfoEntriesFromConfig(configFilePath string) map[string]Entry {
	res := make(map[string]Entry)

	// 1: unmarshal user provided config into boot config struct
	config := &BootConfigAppInfo{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: init rk entry from config
	entry := RegisterAppInfoEntry(
		WithAppNameAppInfo(config.RK.AppName),
		WithVersionAppInfo(config.RK.Version),
		WithDescriptionAppInfo(config.RK.Description),
		WithKeywordsAppInfo(config.RK.Keywords...),
		WithHomeUrlAppInfo(config.RK.HomeUrl),
		WithIconUrlAppInfo(config.RK.IconUrl),
		WithDocsUrlAppInfo(config.RK.DocsUrl...),
		WithMaintainersAppInfo(config.RK.Maintainers...))

	res[AppInfoEntryName] = entry

	return res
}

// Register RKEntry with options.
// This function is used while creating entry from code instead of config file.
// We will override RKEntry fields if value is nil or empty if necessary.
//
// Generally, we recommend call rkctx.GlobalAppCtx.AddEntry() inside this function,
// however, we recommend to register RKEntry, ZapLoggerEntry, EventLoggerEntry with
// function of rkctx.RegisterBasicEntriesWithConfig which will register these entries to
// global context automatically.
func RegisterAppInfoEntry(opts ...AppInfoEntryOption) *AppInfoEntry {
	entry := &AppInfoEntry{
		EntryName:        AppInfoEntryName,
		EntryType:        AppInfoEntryType,
		EntryDescription: AppInfoEntryDescription,
		AppName:          AppNameDefault,
		Version:          VersionDefault,
		Lang:             LangDefault,
		Description:      DescriptionDefault,
		Keywords:         []string{},
		HomeUrl:          "",
		IconUrl:          "",
		DocsUrl:          []string{},
		Maintainers:      []string{},
	}

	for i := range opts {
		opts[i](entry)
	}

	// override elements which should not be nil
	if len(entry.Keywords) < 1 {
		entry.Keywords = []string{}
	}

	if len(entry.DocsUrl) < 1 {
		entry.DocsUrl = []string{}
	}

	if len(entry.Maintainers) < 1 {
		entry.Maintainers = []string{}
	}

	// override elements which should not be empty
	if len(entry.AppName) < 1 {
		entry.AppName = AppNameDefault
	}

	if len(entry.Version) < 1 {
		entry.Version = VersionDefault
	}

	if len(entry.Lang) < 1 {
		entry.Lang = LangDefault
	}

	if len(entry.Description) < 1 {
		entry.Description = DescriptionDefault
	}

	GlobalAppCtx.SetAppInfoEntry(entry)

	// override default event logger entry in order to use correct application name.
	// this is special case for default event logger entry.
	eventLoggerConfig := GlobalAppCtx.GetEventLoggerEntryDefault().LoggerConfig
	eventLogger, _ := eventLoggerConfig.Build()
	eventLoggerEntry := RegisterEventLoggerEntry(
		WithNameEvent(DefaultEventLoggerEntryName),
		WithEventFactoryEvent(
			rkquery.NewEventFactory(
				rkquery.WithLogger(eventLogger),
				rkquery.WithAppName(entry.AppName))))

	eventLoggerEntry.LoggerConfig = eventLoggerConfig

	return entry
}

// No op
func (entry *AppInfoEntry) Bootstrap(context.Context) {
	// no op
}

// No op
func (entry *AppInfoEntry) Interrupt(context.Context) {
	// no op
}

// Return name of entry
func (entry *AppInfoEntry) GetName() string {
	return entry.EntryName
}

// Return type of entry
func (entry *AppInfoEntry) GetType() string {
	return entry.EntryType
}

// Return description of entry
func (entry *AppInfoEntry) GetDescription() string {
	return entry.EntryDescription
}

// Return string of entry
func (entry *AppInfoEntry) String() string {
	if bytes, err := json.Marshal(entry); err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}
