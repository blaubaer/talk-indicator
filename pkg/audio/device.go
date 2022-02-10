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

func (this Device) HasSession() bool {
	return this.Sessions.HasContent()
}

type Devices []Device

func (this Devices) IsZero() bool {
	return len(this) <= 0
}

func (this Devices) HasContent() bool {
	return !this.IsZero()
}

func (this Devices) HasSession() bool {
	for _, v := range this {
		if v.HasSession() {
			return true
		}
	}
	return false
}
