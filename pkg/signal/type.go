package signal

import (
	"fmt"
	"strings"
)

type Type uint8

const (
	TypeHue Type = iota
	TypeHomeAssistant
	TypeSystray

	TypeDefault = TypeHue
)

var (
	AllTypes = Types{
		TypeHue,
		TypeHomeAssistant,
		TypeSystray,
	}
)

func (this *Type) Set(plain string) error {
	switch strings.TrimSpace(strings.ToLower(plain)) {
	case "hue":
		*this = TypeHue
		return nil
	case "homeassistant":
		*this = TypeHomeAssistant
		return nil
	case "systray":
		*this = TypeSystray
		return nil
	default:
		return fmt.Errorf("illegal-signal-type: %s", plain)
	}
}

func (this Type) String() string {
	v, err := this.MarshalText()
	if err != nil {
		return fmt.Sprintf("illegal-signal-type-%d", this)
	}
	return string(v)
}

func (this Type) MarshalText() (text []byte, err error) {
	switch this {
	case TypeHue:
		return []byte("hue"), nil
	case TypeHomeAssistant:
		return []byte("homeassistant"), nil
	case TypeSystray:
		return []byte("systray"), nil
	default:
		return nil, fmt.Errorf("illegal signal type: %v", this)
	}
}

func (this *Type) UnmarshalText(text []byte) error {
	return this.Set(string(text))
}

type Types []Type

func (this Types) Strings() []string {
	result := make([]string, len(this))
	for i, v := range this {
		result[i] = v.String()
	}
	return result
}

func (this Types) String() string {
	return strings.Join(this.Strings(), ",")
}
