//go:build windows

package app

import (
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/signal/facade"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

func NewConfiguration() Configuration {
	return Configuration{
		false,

		facade.NewConfiguration(),

		5 * time.Second,
		5 * time.Minute,

		common.Regexp{},
		common.MustNewRegexp(`\{[0-9a-f.]+}\.{[0-9a-f-]+}\|\\Device\\.+\\Windows\\System32\\svchost\.exe%.*`),
	}
}

type Configuration struct {
	PreventAutoSave bool `yaml:"preventAutoSave"`

	Signal facade.Configuration `yaml:"signal,omitempty"`

	CheckInterval   time.Duration `yaml:"checkInterval,omitempty"`
	RefreshInterval time.Duration `yaml:"refreshInterval,omitempty"`

	IncludedSessionIdentifiers common.Regexp `yaml:"includedSessionIdentifiers,omitempty"`
	ExcludedSessionIdentifiers common.Regexp `yaml:"excludedSessionIdentifiers,omitempty"`
}

func (this *Configuration) SetupConfiguration(using common.FlagHolder) {
	using.Flag("preventAutoSave", "If provided configuration will NOT automatically be saved upon changes.").
		Envar("TI_PREVENT_AUTO_SAVE").
		BoolVar(&this.PreventAutoSave)
	using.Flag("checkInterval", "How often the state of the talk is checked.").
		Envar("TI_CHECK_INTERVAL").
		DurationVar(&this.CheckInterval)
	using.Flag("refreshInterval", "How often the whole setup should be refreshed.").
		Envar("TI_REFRESH_INTERVAL").
		DurationVar(&this.RefreshInterval)
	using.Flag("includedSessionIdentifiers", "Which session identifiers should be respected for evaluation.").
		Envar("TI_INCLUDED_SESSION_IDENTIFIERS").
		SetValue(&this.IncludedSessionIdentifiers)
	using.Flag("excludedSessionIdentifiers", "Which session identifiers should not be respected for evaluation.").
		Envar("TI_EXCLUDED_SESSION_IDENTIFIERS").
		SetValue(&this.ExcludedSessionIdentifiers)
}

func defaultConfigurationFile() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		fs, err := os.Stat(appData)
		if err == nil && fs.IsDir() {
			return filepath.Join(appData, "talk-indicator", "configuration.yml")
		}
	}

	u, err := user.Current()
	if err != nil {
		return "configuration.yaml"
	}

	return filepath.Join(u.HomeDir, ".config", "talk-indicator", "configuration.yml")
}
