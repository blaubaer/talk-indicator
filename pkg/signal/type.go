package signal

import (
	"fmt"
	"strings"
)

type Type uint8

const (
	TypeHue Type = iota
	TypeSystray

	TypeDefault = TypeHue
)

var (
	AllTypes = Types{
		TypeHue,
		TypeSystray,
	}
)

func (this *Type) Set(plain string) error {
	switch strings.TrimSpace(strings.ToLower(plain)) {
	case "hue":
		*this = TypeHue
		return nil
	case "systray":
		*this = TypeSystray
		return nil
	default:
		return fmt.Errorf("illegal-signal-type: %s", plain)
	}
}

func (this Type) String() string {
	switch this {
	case TypeHue:
		return "hue"
	case TypeSystray:
		return "systray"
	default:
		return fmt.Sprintf("illegal-signal-type-%d", this)
	}
}

func (this Type) newInstance() Signal {
	switch this {
	case TypeHue:
		return &Hue{}
	case TypeSystray:
		return &Systray{}
	default:
		panic(fmt.Errorf("illegal-signal-type-%d", this))
	}
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
