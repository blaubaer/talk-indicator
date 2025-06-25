package signal

import (
	"fmt"
	"strings"
)

type State uint8

const (
	StateOff = State(0)
	StateOn  = State(1)
)

var (
	AllStates = States{
		StateOff,
		StateOn,
	}
)

func (this *State) Set(plain string) error {
	switch strings.TrimSpace(strings.ToLower(plain)) {
	case "off", "0", "false", "no":
		*this = StateOff
		return nil
	case "on", "1", "true", "yes":
		*this = StateOn
		return nil
	default:
		return fmt.Errorf("illegal-signal-state: %s", plain)
	}
}

func (this State) String() string {
	v, err := this.MarshalText()
	if err != nil {
		return fmt.Sprintf("illegal-signal-state-%d", this)
	}
	return string(v)
}

func (this State) MarshalText() (text []byte, err error) {
	switch this {
	case StateOff:
		return []byte("off"), nil
	case StateOn:
		return []byte("on"), nil
	default:
		return nil, fmt.Errorf("illegal signal state: %v", this)
	}
}

func (this *State) UnmarshalText(text []byte) error {
	return this.Set(string(text))
}

type States []State

func (this States) Strings() []string {
	result := make([]string, len(this))
	for i, v := range this {
		result[i] = v.String()
	}
	return result
}

func (this States) String() string {
	return strings.Join(this.Strings(), ",")
}
