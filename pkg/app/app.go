package app

import (
	"context"
	"github.com/blaubaer/talk-indicator/pkg/audio"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	log "github.com/echocat/slf4g"
	"regexp"
	"sync"
	"time"
)

type App struct {
	AudioStack audio.Stack
	Signal     signal.Facade

	CheckInterval   time.Duration
	RefreshInterval time.Duration

	IncludedSessionIdentifiers *regexp.Regexp
	ExcludedSessionIdentifiers *regexp.Regexp

	initialized sync.Once
}

func (this *App) ensure() {
	this.initialized.Do(func() {
		this.CheckInterval = 5 * time.Second
		this.RefreshInterval = 5 * time.Minute
		this.ExcludedSessionIdentifiers = regexp.MustCompile(`\{[0-9a-f.]+}\.{[0-9a-f-]+}\|\\Device\\.+\\Windows\\System32\\svchost\.exe%.*`)
	})
}

func (this *App) SetupConfiguration(using common.FlagHolder) {
	this.ensure()

	this.AudioStack.SetupConfiguration(using)
	this.Signal.SetupConfiguration(using)

	var includedSessionIdsDef, excludedSessionIdsDef string
	if v := this.IncludedSessionIdentifiers; v != nil {
		includedSessionIdsDef = v.String()
	}
	if v := this.ExcludedSessionIdentifiers; v != nil {
		excludedSessionIdsDef = v.String()
	}

	using.Flag("checkInterval", "How often the state of the talk is checked.").
		Envar("TI_CHECK_INTERVAL").
		Default(this.CheckInterval.String()).
		DurationVar(&this.CheckInterval)
	using.Flag("refreshInterval", "How often the whole setup should be refreshed.").
		Envar("TI_REFRESH_INTERVAL").
		Default(this.RefreshInterval.String()).
		DurationVar(&this.RefreshInterval)
	using.Flag("includedSessionIdentifiers", "Which session identifiers should be respected for evaluation.").
		Envar("TI_INCLUDED_SESSION_IDENTIFIERS").
		Default(includedSessionIdsDef).
		RegexpVar(&this.IncludedSessionIdentifiers)
	using.Flag("excludedSessionIdentifiers", "Which session identifiers should not be respected for evaluation.").
		Envar("TI_EXCLUDED_SESSION_IDENTIFIERS").
		Default(excludedSessionIdsDef).
		RegexpVar(&this.ExcludedSessionIdentifiers)
}

func (this *App) Run(ctx context.Context) error {
	this.ensure()

	var lastState *signal.State

	ctxInner, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			log.With("interval", this.RefreshInterval).
				Debug("Wait until the next refresh...")
			select {
			case <-ctxInner.Done():
				log.Debug("Refresh loop interrupted.")
				return
			case <-time.After(this.RefreshInterval):
			}

			if err := this.Signal.Update(); err != nil {
				log.WithError(err).
					Error("Cannot update signal.")
				continue
			}

			if lastState != nil {
				if err := this.Signal.Ensure(*lastState); err != nil {
					log.WithError(err).
						Error("Cannot ensure signal state.")
					continue
				}
			}
		}
	}()

	first := true
	for {
		if first {
			first = false
		} else {
			log.With("interval", this.CheckInterval).
				Debug("Wait until the next check...")
			select {
			case <-ctx.Done():
				log.Debug("Check loop interrupted.")
				return nil
			case <-time.After(this.CheckInterval):
			}
		}

		devices, err := this.AudioStack.FindDevices()
		if err != nil {
			log.WithError(err).
				Error("Cannot find audio devices.")
			continue
		}

		state := signal.StateOff
		if devices.HasRelevantSession(func(candidate *audio.Session) bool {
			if v := this.IncludedSessionIdentifiers; v != nil && v.String() != "" {
				if !v.MatchString(candidate.Identifier) {
					return false
				}
			}
			if v := this.ExcludedSessionIdentifiers; v != nil && v.String() != "" {
				if v.MatchString(candidate.Identifier) {
					return false
				}
			}
			return true
		}) {
			state = signal.StateOn
		}

		log.With("devices", devices).
			With("state", state).
			Debug("Devices and their sessions discovered.")

		if lastState == nil || *lastState != state {
			if lastState == nil {
				buf := signal.StateOff
				lastState = &buf
			}
			log.With("lastState", *lastState).
				With("state", state).
				Info("State change detected.")
		}

		if err := this.Signal.Ensure(state); err != nil {
			log.WithError(err).
				Error("It was not possible to ensure signal state.")
			continue
		}

		lastState = &state
	}
}

func (this *App) Initialize() (rErr error) {
	this.ensure()

	success := false
	defer func() {
		if !success {
			if err := this.Dispose(); err != nil && rErr == nil {
				rErr = err
			}
		}
	}()

	if err := this.AudioStack.Initialize(); err != nil {
		return err
	}
	if err := this.Signal.Initialize(); err != nil {
		return err
	}

	success = true
	return nil
}

func (this *App) Dispose() (rErr error) {
	this.ensure()

	defer func() {
		if err := this.AudioStack.Dispose(); err != nil && rErr == nil {
			rErr = err
		}
	}()

	defer func() {
		if err := this.Signal.Dispose(); err != nil && rErr == nil {
			rErr = err
		}
	}()

	return this.Signal.Ensure(signal.StateOff)
}
