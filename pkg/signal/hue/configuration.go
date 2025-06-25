package hue

import "github.com/blaubaer/talk-indicator/pkg/common"

func NewConfiguration() Configuration {
	return Configuration{
		false,
		"",
		"",

		common.MustNewRegexp("^OnAir"),
		HueKinds{},

		254,
		65535,
		254,
	}
}

type Configuration struct {
	Pair   bool   `yaml:"pair,omitempty"`
	Bridge string `yaml:"bridge,omitempty"`
	User   string `yaml:"user,omitempty"`

	Name  common.Regexp `yaml:"target"`
	Kinds HueKinds      `yaml:"kinds,omitempty"`

	Brightness uint8  `yaml:"brightness"`
	Hue        uint16 `yaml:"hue"`
	Saturation uint8  `yaml:"saturation"`
}

func (this *Configuration) SetupConfiguration(using common.FlagHolder) {
	using.Flag("signal.hue.pair", "If true this application will pair again with an existing hue. This will be implicit enabled if this application is not already paired.").
		Envar("TI_SIGNAL_HUE_PAIR").
		BoolVar(&this.Pair)
	using.Flag("signal.hue.bridge", "Usually the bridge is automatically detected. You can specify an explicit one if they are more than one. This is only required while pairing and will afterwards be ignored.").
		Envar("TI_SIGNAL_HUE_BRIDGE").
		StringVar(&this.Bridge)
	using.Flag("signal.hue.user", "Usually this is set while pairing and will then be persisted. If this set this will be used and not be persisted.").
		Envar("TI_SIGNAL_HUE_USER").
		StringVar(&this.User)
	using.Flag("signal.hue.name", "Name as regex of the lights/groups which should be handled by this app.").
		Envar("TI_SIGNAL_HUE_NAME").
		SetValue(&this.Name)
	using.Flag("signal.hue.kind", "Kind(s) of what should be handled. Possible values: "+AllHueKinds.String()).
		Envar("TI_SIGNAL_HUE_KIND").
		SetValue(&this.Kinds)

	using.Flag("signal.hue.brightness", "The brightness value to set the light to. Brightness is a scale from 1 (the minimum the light is capable of) to 254 (the maximum).").
		Envar("TI_SIGNAL_HUE_BRIGHTNESS").
		Uint8Var(&this.Brightness)
	using.Flag("signal.hue.hue", "The hue value to set light to. The hue value is a wrapping value between 0 and 65535. Both 0 and 65535 are red, 25500 is green and 46920 is blue.").
		Envar("TI_SIGNAL_HUE_HUE").
		Uint16Var(&this.Hue)
	using.Flag("signal.hue.saturation", "Saturation of the light. 254 is the most saturated (colored) and 0 is the least saturated (white).").
		Envar("TI_SIGNAL_HUE_SATURATION").
		Uint8Var(&this.Saturation)
}
