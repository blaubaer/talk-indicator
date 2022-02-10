package signal

import (
	"fmt"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"sync"
)

type Facade struct {
	Signal

	initialized sync.Once
	typeFacade  facadeTypeFacade
}

func (this *Facade) SetupConfiguration(using common.FlagHolder) {
	this.ensure()
	this.typeFacade.SetupConfiguration(using)
}

func (this *Facade) Initialize() error {
	this.ensure()
	return this.Signal.Initialize()
}

func (this *Facade) Dispose() error {
	this.ensure()
	return this.Signal.Dispose()
}

func (this *Facade) GetType() Type {
	this.ensure()
	return this.Signal.GetType()
}

func (this *Facade) ensure() {
	this.initialized.Do(func() {
		this.typeFacade.owner = this
		this.typeFacade.allVariants = make(map[Type]Signal, len(AllTypes))
		for _, t := range AllTypes {
			this.typeFacade.allVariants[t] = t.newInstance()
		}
		this.Signal = this.typeFacade.allVariants[TypeDefault]
	})
}

type facadeTypeFacade struct {
	owner       *Facade
	allVariants map[Type]Signal
}

func (this *facadeTypeFacade) Set(plain string) error {
	var t Type
	if err := t.Set(plain); err != nil {
		return err
	}
	s, ok := this.allVariants[t]
	if !ok {
		return fmt.Errorf("illegal-signal-tyle: %s", plain)
	}
	this.owner.Signal = s
	return nil
}

func (this *facadeTypeFacade) String() string {
	return this.owner.Signal.GetType().String()
}

func (this *facadeTypeFacade) SetupConfiguration(using common.FlagHolder) {
	using.Flag("signal.type", fmt.Sprintf("Type how the signal should be sent. Possible values: %v", AllTypes)).
		Default(TypeDefault.String()).
		Envar("TI_SIGNAL_TYPE").
		SetValue(this)

	for _, s := range this.allVariants {
		s.SetupConfiguration(using)
	}
}
