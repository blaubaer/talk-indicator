package systray

import (
	"fmt"
	"github.com/blaubaer/talk-indicator/pkg/signal"
	"strings"

	"github.com/getlantern/systray"
)

type Systray struct {
	IconOn  []byte
	IconOff []byte
}

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

func (this *Systray) Ensure(ctx signal.Context) error {
	state := ctx.State()
	if state == signal.StateOff {
		systray.SetIcon(this.IconOff)
		systray.SetTooltip("Nobody is using the microphone")
		return nil
	}

	var sessionStr []string
	for session, err := range ctx.Sessions() {
		if err != nil {
			return err
		}
		sessionStr = append(sessionStr, session.Title)
	}
	systray.SetIcon(this.IconOn)
	systray.SetTooltip(fmt.Sprintf("Microphone is used by:\n%s", strings.Join(sessionStr, "\n")))

	return nil
}

func (this *Systray) Update() error {
	return nil
}

func (this *Systray) GetType() signal.Type {
	return signal.TypeSystray
}
