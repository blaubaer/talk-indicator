package facade

import (
	"fmt"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	"github.com/blaubaer/talk-indicator/pkg/signal/homeassistant"
	"github.com/blaubaer/talk-indicator/pkg/signal/hue"
	"github.com/blaubaer/talk-indicator/pkg/signal/systray"
	"sync"
)

type Facade struct {
	signal.Signal

	lock sync.RWMutex
}

func (this *Facade) Ensure(c signal.Context) error {
	this.lock.RLock()
	defer this.lock.RUnlock()

	if v := this.Signal; v != nil {
		return v.Ensure(c)
	}
	return nil
}

func (this *Facade) Update() error {
	this.lock.RLock()
	defer this.lock.RUnlock()

	if v := this.Signal; v != nil {
		return v.Update()
	}
	return nil
}

func (this *Facade) Initialize(conf *Configuration, saveConfFunc func() error) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.Signal != nil {
		return nil
	}

	switch conf.Type {
	case signal.TypeHue:
		var buf hue.Hue
		if err := buf.Initialize(&conf.Hue, saveConfFunc); err != nil {
			return err
		}
		this.Signal = &buf
	case signal.TypeHomeAssistant:
		var buf homeassistant.Homeassistant
		if err := buf.Initialize(&conf.HomeAssistant, saveConfFunc); err != nil {
			return err
		}
		this.Signal = &buf
	case signal.TypeSystray:
		var buf systray.Systray
		if err := buf.Initialize(); err != nil {
			return err
		}
		this.Signal = &buf
	default:
		return fmt.Errorf("unsupported signal type: %v", conf.Type)
	}

	return nil
}

func (this *Facade) Dispose() error {
	this.lock.Lock()
	defer this.lock.Unlock()

	defer func() {
		this.Signal = nil
	}()

	if v := this.Signal; v != nil {
		return v.Dispose()
	}
	return nil
}

func (this *Facade) GetType() signal.Type {
	this.lock.RLock()
	defer this.lock.RUnlock()

	if v := this.Signal; v != nil {
		return v.GetType()
	}

	return 0
}
