package app

import (
	"context"
	"github.com/blaubaer/talk-indicator/pkg/audio"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	log "github.com/echocat/slf4g"
	"sync"
	"time"
)

type App struct {
	AudioStack audio.Stack
	Signal     signal.Facade

	CheckInterval   time.Duration
	RefreshInterval time.Duration

	initialized sync.Once
}

func (this *App) ensure() {
	this.initialized.Do(func() {
		this.CheckInterval = 5 * time.Second
		this.RefreshInterval = 5 * time.Minute
	})
}

func (this *App) SetupConfiguration(using common.FlagHolder) {
	this.ensure()

	this.AudioStack.SetupConfiguration(using)
	this.Signal.SetupConfiguration(using)

	using.Flag("checkInterval", "How often the state of the talk is checked.").
		Envar("TI_CHECK_INTERVAL").
		Default(this.CheckInterval.String()).
		DurationVar(&this.CheckInterval)
	using.Flag("refreshInterval", "How often the whole setup should be refreshed.").
		Envar("TI_REFRESH_INTERVAL").
		Default(this.RefreshInterval.String()).
		DurationVar(&this.RefreshInterval)
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
		if devices.HasSession() {
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
