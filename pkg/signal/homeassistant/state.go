package homeassistant

import (
	"encoding/json"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	"time"
)

type stateGetResponse struct {
	EntityId     string         `json:"entity_id"`
	State        signal.State   `json:"state"`
	Attributes   map[string]any `json:"attributes"`
	LastChanged  time.Time      `json:"last_changed"`
	LastReported time.Time      `json:"last_reported"`
	LastUpdated  time.Time      `json:"last_updated"`
	Context      map[string]any `json:"context"`
}

func (this *stateGetResponse) getAttrSessions() (stateAttrSessions, error) {
	var result stateAttrSessions
	if this.Attributes != nil {
		if plain, ok := this.Attributes["sessions"]; ok {
			if err := result.unmarshalFromAny(plain); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

type statePostRequest struct {
	State      signal.State   `json:"state"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (this *statePostRequest) setAttrSessions(v stateAttrSessions) {
	if this.Attributes == nil {
		this.Attributes = make(map[string]any)
	}
	this.Attributes["sessions"] = v
}

type stateAttrSessions []stateAttrSession

func (this *stateAttrSessions) unmarshalFromAny(in any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, this)
}

func (this stateAttrSessions) isEqualTo(o *stateAttrSessions) bool {
	if len(this) != len(*o) {
		return false
	}
	for i, tV := range this {
		oV := (*o)[i]
		if !tV.isEqualTo(&oV) {
			return false
		}
	}
	return true
}

type stateAttrSession struct {
	Title      string `json:"title"`
	Device     string `json:"device"`
	Identifier string `json:"identifier,omitempty"`
}

func (this stateAttrSession) isEqualTo(o *stateAttrSession) bool {
	return this.Title == o.Title &&
		this.Device == o.Device &&
		this.Identifier == o.Identifier
}
