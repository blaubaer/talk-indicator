package app

import (
	"context"
	"dario.cat/mergo"
	"github.com/blaubaer/talk-indicator/pkg/audio"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	"github.com/blaubaer/talk-indicator/pkg/signal/facade"
	log "github.com/echocat/slf4g"
	"iter"
	"os"
	"sync"
	"time"
)

func NewApp() *App {
	return &App{
		config: NewConfiguration(),
	}
}

type App struct {
	AudioStack        audio.Stack
	Signal            facade.Facade
	OtherSignals      []signal.Signal
	ConfigurationFile string

	configFromFlags Configuration
	config          Configuration
	initialized     sync.Once
}

func (this *App) SetupConfiguration(using common.FlagHolder) {
	this.AudioStack.SetupConfiguration(using)
	this.configFromFlags.SetupConfiguration(using)

	using.Flag("configuration", "Defines the file from which the configuration should be loaded and/or stored to.").
		Short('c').
		StringVar(&this.ConfigurationFile)
}

func (this *App) Run(ctx context.Context) error {
	var lastState *signal.State
	var lastSessions audio.Sessions

	ctxInner, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			log.With("interval", this.config.RefreshInterval).
				Debug("Wait until the next refresh...")
			select {
			case <-ctxInner.Done():
				log.Debug("Refresh loop interrupted.")
				return
			case <-time.After(this.config.RefreshInterval):
			}

			if err := this.Signal.Update(); err != nil {
				log.WithError(err).
					Error("Cannot update signal.")
				continue
			}
			for _, s := range this.OtherSignals {
				if err := s.Update(); err != nil {
					log.WithError(err).
						Warn("Cannot update signal.")
				}
			}

			if lastState != nil {
				sCtx := &signalContext{this, &lastSessions, *lastState}
				if err := this.Signal.Ensure(sCtx); err != nil {
					log.WithError(err).
						Error("Cannot ensure signal state.")
					continue
				}
				for _, s := range this.OtherSignals {
					if err := s.Ensure(sCtx); err != nil {
						log.WithError(err).
							Warn("Cannot ensure signal state.")
					}
				}

			}
		}
	}()

	first := true
	for {
		if first {
			first = false
		} else {
			log.With("interval", this.config.CheckInterval).
				Debug("Wait until the next check...")
			select {
			case <-ctx.Done():
				log.Debug("Check loop interrupted.")
				return nil
			case <-time.After(this.config.CheckInterval):
			}
		}

		allDevices, err := this.AudioStack.FindDevices()
		if err != nil {
			log.WithError(err).
				Error("Cannot find audio devices.")
			continue
		}

		lastSessions, err = audio.CollectSessionsErr(allDevices.RelevantSessions(this.isSessionRelevant))
		if err != nil {
			log.WithError(err).
				Error("Cannot evaluate audio sessions.")
			continue
		}

		state := signal.StateOff
		if len(lastSessions) > 0 {
			state = signal.StateOn
		}

		log.With("devices", allDevices).
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

		if err := this.Signal.Ensure(&signalContext{this, &lastSessions, state}); err != nil {
			log.WithError(err).
				Error("It was not possible to ensure signal state.")
			continue
		}
		for _, s := range this.OtherSignals {
			if err := s.Ensure(&signalContext{this, &lastSessions, state}); err != nil {
				log.WithError(err).
					Warn("It was not possible to ensure signal state.")
			}
		}
		lastState = &state
	}
}

type signalContext struct {
	app   *App
	sess  *audio.Sessions
	state signal.State
}

func (this *signalContext) State() signal.State {
	return this.state
}

func (this *signalContext) Sessions() iter.Seq2[*audio.Session, error] {
	return func(yield func(*audio.Session, error) bool) {
		vs := this.sess
		if vs == nil {
			return
		}
		for _, v := range *vs {
			if !yield(&v, nil) {
				return
			}
		}
	}
}

func (this *App) isSessionRelevant(candidate *audio.Session) (bool, error) {
	if v := this.config.IncludedSessionIdentifiers; v.HasContent() {
		if !v.MatchString(candidate.Identifier) {
			return false, nil
		}
	}
	if v := this.config.ExcludedSessionIdentifiers; v.HasContent() {
		if v.MatchString(candidate.Identifier) {
			return false, nil
		}
	}
	return true, nil
}

func (this *App) Initialize() (rErr error) {
	success := false
	defer func() {
		if !success {
			if err := this.Dispose(); err != nil && rErr == nil {
				rErr = err
			}
		}
	}()

	if err := this.config.loadDefault(true); err != nil {
		return err
	}
	if err := mergo.Merge(&this.config, this.configFromFlags); err != nil {
		return err
	}

	if err := this.AudioStack.Initialize(); err != nil {
		return err
	}
	if err := this.Signal.Initialize(&this.config.Signal, this.alwaysSaveConf); err != nil {
		return err
	}

	if err := this.saveConf(false); err != nil {
		return err
	}

	success = true
	return nil
}

func (this *App) alwaysSaveConf() error {
	return this.saveConf(true)
}

func (this *App) saveConf(always bool) error {
	if this.config.PreventAutoSave {
		log.Debug("Automatically save of configuration disabled.")
		return nil
	}

	fn := defaultConfigurationFile()
	if !always {
		_, err := os.Stat(fn)
		if os.IsNotExist(err) {
			log.With("file", fn).Info("Configuration absent.")
			// Ok, we should save...
		} else if err != nil {
			return err
		} else {
			// Does exist, skip...
			return nil
		}
	}

	if err := this.config.saveToFile(fn); err != nil {
		return err
	}

	log.With("file", fn).Info("Configuration saved.")

	return nil
}

func (this *App) Dispose() (rErr error) {
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

	sCtx := signalContext{this, nil, signal.StateOff}

	for _, s := range this.OtherSignals {
		defer func() { _ = s.Ensure(&sCtx) }()
	}

	return this.Signal.Ensure(&sCtx)
}
