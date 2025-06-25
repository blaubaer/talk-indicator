package signal

import (
	"fmt"
	"strings"

	"github.com/getlantern/systray"

	"github.com/blaubaer/talk-indicator/pkg/audio"
	"github.com/blaubaer/talk-indicator/pkg/common"
)

type Systray struct {
	IconOn  []byte
	IconOff []byte
}

func (this *Systray) SetupConfiguration(holder common.FlagHolder) {}

func (this *Systray) Initialize() error {
	if len(this.IconOn) == 0 {
		return fmt.Errorf("IconOn is empty")
	}
	if len(this.IconOff) == 0 {
		return fmt.Errorf("IconOn is empty")
	}
	return nil
}

func (this *Systray) Dispose() error {
	return nil
}

func (this *Systray) Ensure(state State, devices []audio.Device) error {
	if state == StateOff {
		systray.SetIcon(this.IconOff)
		systray.SetTooltip("Nobody is using the microphone")
		return nil
	}

	var sessionStr []string
	for _, device := range devices {
		for _, session := range device.Sessions {
			sessionStr = append(sessionStr, session.Title)
		}
	}
	systray.SetIcon(this.IconOn)
	systray.SetTooltip(fmt.Sprintf("Microphone is used by:\n%s", strings.Join(sessionStr, "\n")))

	return nil
}

func (this *Systray) Update() error {
	return nil
}

func (this *Systray) GetType() Type {
	return TypeSystray
}
