package facade

import (
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	"github.com/blaubaer/talk-indicator/pkg/signal/homeassistant"
	"github.com/blaubaer/talk-indicator/pkg/signal/hue"
)

func NewConfiguration() Configuration {
	return Configuration{
		Type:          signal.TypeDefault,
		Hue:           hue.NewConfiguration(),
		HomeAssistant: homeassistant.NewConfiguration(),
	}
}

type Configuration struct {
	Type          signal.Type                 `yaml:"type"`
	Hue           hue.Configuration           `yaml:"hue,omitempty"`
	HomeAssistant homeassistant.Configuration `yaml:"homeAssistant,omitempty"`
}

func (this *Configuration) SetupConfiguration(using common.FlagHolder) {
	using.Flag("signal", "Signal to use. All possible values: "+signal.AllTypes.String()).
		Envar("TI_SIGNAL").
		SetValue(&this.Type)

	this.Hue.SetupConfiguration(using)
}
