package audio

import "github.com/blaubaer/talk-indicator/pkg/common"

type Stack struct{}

func (this *Stack) SetupConfiguration(_ common.FlagHolder) {}

func (this *Stack) Initialize() error {
	return nil
}

func (this *Stack) Dispose() error {
	return nil
}

func (this *Stack) FindDevices() (Devices, error) {
	return findDevices()
}
