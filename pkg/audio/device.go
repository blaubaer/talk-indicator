package audio

import (
	"fmt"
)

type Device struct {
	Name     string   `json:"name"`
	Index    uint32   `json:"index"`
	Sessions Sessions `json:"sessions,omitempty"`
}

func (this Device) String() string {
	return fmt.Sprintf("[%d] %s", this.Index, this.Name)
}

func (this Device) HasRelevantSession(predicate func(*Session) bool) bool {
	return this.Sessions.HasRelevantSession(predicate)
}

type Devices []Device

func (this Devices) IsZero() bool {
	return len(this) <= 0
}

func (this Devices) HasContent() bool {
	return !this.IsZero()
}

func (this Devices) HasRelevantSession(predicate func(*Session) bool) bool {
	for _, v := range this {
		if v.HasRelevantSession(predicate) {
			return true
		}
	}
	return false
}
