package signal

import (
	"fmt"
	"strings"
)

type HueKind uint8

const (
	HueKindLight = HueKind(0)
	HueKindGroup = HueKind(1)
)

var (
	AllHueKinds = HueKinds{
		HueKindLight,
		HueKindGroup,
	}
)

func (this *HueKind) Set(plain string) error {
	switch strings.TrimSpace(strings.ToLower(plain)) {
	case "light":
		*this = HueKindLight
		return nil
	case "group", "room":
		*this = HueKindGroup
		return nil
	default:
		return fmt.Errorf("illegal-signal-hue-kind: %s", plain)
	}
}

func (this HueKind) String() string {
	switch this {
	case HueKindLight:
		return "light"
	case HueKindGroup:
		return "group"
	default:
		return fmt.Sprintf("illegal-signal-hue-kind-%d", this)
	}
}

type HueKinds []HueKind

func (this *HueKinds) Set(plain string) error {
	for _, plain := range strings.Split(plain, ",") {
		plain = strings.TrimSpace(plain)
		if plain != "" {
			var v HueKind
			if err := v.Set(plain); err != nil {
				return err
			}
			*this = append(*this, v)
		}
	}
	return nil
}

func (this HueKinds) Strings() []string {
	result := make([]string, len(this))
	for i, v := range this {
		result[i] = v.String()
	}
	return result
}

func (this HueKinds) String() string {
	return strings.Join(this.Strings(), ",")
}

func (this HueKinds) IsCumulative() bool {
	return true
}

func (this HueKinds) Has(v HueKind) bool {
	if len(this) == 0 {
		return true
	}
	for _, candidate := range this {
		if v == candidate {
			return true
		}
	}
	return false
}
