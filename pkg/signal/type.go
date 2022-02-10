package signal

import (
	"fmt"
	"strings"
)

type Type uint8

const (
	TypeHue = Type(0)

	TypeDefault = TypeHue
)

var (
	AllTypes = Types{
		TypeHue,
	}
)

func (this *Type) Set(plain string) error {
	switch strings.TrimSpace(strings.ToLower(plain)) {
	case "hue":
		*this = TypeHue
		return nil
	default:
		return fmt.Errorf("illegal-signal-type: %s", plain)
	}
}

func (this Type) String() string {
	switch this {
	case TypeHue:
		return "hue"
	default:
		return fmt.Sprintf("illegal-signal-type-%d", this)
	}
}

func (this Type) newInstance() Signal {
	switch this {
	case TypeHue:
		return &Hue{}
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
