package signal

import (
	"encoding/json"
	"github.com/amimof/huego"
)

type HueCredentials struct {
	Host string `json:"host"`
	User string `json:"user"`
}

func (this HueCredentials) IsZero() bool {
	return this.Host == "" || this.User == ""
}

func (this HueCredentials) HasContent() bool {
	return !this.IsZero()
}

func (this HueCredentials) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *HueCredentials) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}

func (this HueCredentials) Bridge() *huego.Bridge {
	return huego.New(this.Host, this.User)
}
