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
	switch this {
	case StateOff:
		return "off"
	case StateOn:
		return "on"
	default:
		return fmt.Sprintf("illegal-signal-state-%d", this)
	}
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
