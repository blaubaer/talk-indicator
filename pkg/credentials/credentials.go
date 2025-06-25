package credentials

import (
	"encoding/json"
)

const appName = "github.com/blaubaer/talk-indicator"

type Credentials struct {
	HueBridge string `json:"hue_bridge,omitempty"`
	HueUser   string `json:"hue_user,omitempty"`

	HomeAssistantServer string `json:"homeAssistant_server,omitempty"`
	HomeAssistantToken  string `json:"homeAssistant_token,omitempty"`

	// Deprecated: Use HueBridge from now on.
	Host string `json:"host,omitempty"`
	// Deprecated: Use HueUser from now on.
	User string `json:"user,omitempty"`
}

func (this *Credentials) IsZero() bool {
	return this.IsHueZero() && this.IsHomeAssistantZero()
}

func (this *Credentials) IsHueZero() bool {
	return this.HueBridge == "" && this.HueUser == ""
}

func (this *Credentials) IsHomeAssistantZero() bool {
	return this.HomeAssistantServer == "" && this.HomeAssistantToken == ""
}

func (this *Credentials) MarshalBinary() (data []byte, err error) {
	this.migrate()

	return json.Marshal(this)
}

func (this *Credentials) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, this); err != nil {
		return err
	}

	this.migrate()

	return nil
}

//goland:noinspection GoDeprecation
func (this *Credentials) migrate() {
	if this.Host != "" && this.HueBridge == "" {
		this.HueBridge = this.Host
		this.Host = ""
	}
	if this.User != "" && this.HueUser == "" {
		this.HueUser = this.User
		this.User = ""
	}
}
