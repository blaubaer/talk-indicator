package signal

import (
	"github.com/blaubaer/talk-indicator/pkg/audio"
	"github.com/blaubaer/talk-indicator/pkg/common"
)

type Signal interface {
	SetupConfiguration(common.FlagHolder)
	Initialize() error
	Dispose() error
	Ensure(State, []audio.Device) error
	Update() error

	GetType() Type
}
