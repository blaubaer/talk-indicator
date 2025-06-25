package audio

import (
	"fmt"
	"iter"
	"strings"
)

type Device struct {
	Name     string   `json:"name"`
	Index    uint32   `json:"index"`
	Sessions Sessions `json:"sessions,omitempty"`
}

func (this Device) String() string {
	return fmt.Sprintf("[%d] %s", this.Index, this.Name)
}

func (this Device) CloneBare() Device {
	return Device{
		strings.Clone(this.Name),
		this.Index,
		Sessions{},
	}
}

func (this Device) RelevantSessions(predicate func(*Session) (bool, error)) iter.Seq2[*Session, error] {
	return this.Sessions.RelevantSessions(predicate)
}

type Devices []Device

func (this Devices) IsZero() bool {
	return len(this) <= 0
}

func (this Devices) HasContent() bool {
	return !this.IsZero()
}

func (this Devices) HasRelevantSession(predicate func(*Session) (bool, error)) (bool, error) {
	for _, err := range this.RelevantSessions(predicate) {
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (this Devices) RelevantSessions(predicate func(*Session) (bool, error)) iter.Seq2[*Session, error] {
	return func(yield func(*Session, error) bool) {
		for _, v := range this {
			for sess, err := range v.RelevantSessions(predicate) {
				if !yield(sess, err) {
					return
				}
			}
		}
	}
}
