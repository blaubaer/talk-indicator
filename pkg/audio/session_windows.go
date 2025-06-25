//go:build windows

package audio

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/windows"

	"github.com/blaubaer/talk-indicator/pkg/common"
)

func (this Device) getSessionsOfDevice(sessionManager *wca.IAudioSessionManager2, dev *Device) (result Sessions, _ error) {
	var enumerator *wca.IAudioSessionEnumerator
	if err := sessionManager.GetSessionEnumerator(&enumerator); err != nil {
		return nil, fmt.Errorf("cannot get audio sessions of device %v: %w", this, err)
	}
	defer enumerator.Release()

	var count int
	if err := enumerator.GetCount(&count); err != nil {
		return nil, fmt.Errorf("cannot get count of audio sessions of device %v: %w", this, err)
	}

	for i := 0; i < count; i++ {
		session, ok, err := this.introspectSessionOf(enumerator, i)
		if err != nil {
			return nil, err
		}
		session.Device = dev.CloneBare()
		if ok {
			result = append(result, session)
		}
	}
	return
}

func (this Device) introspectSessionOf(sessions *wca.IAudioSessionEnumerator, sessionIndex int) (Session, bool, error) {
	var sessionControl *wca.IAudioSessionControl
	if err := sessions.GetSession(sessionIndex, &sessionControl); err != nil {
		return Session{}, false, fmt.Errorf("cannot get audio session %d of device %v: %w", sessionIndex, this, err)
	}
	defer sessionControl.Release()

	return this.introspectSession(sessionControl, sessionIndex)
}

func (this Device) introspectSession(sessionControl *wca.IAudioSessionControl, sessionIndex int) (Session, bool, error) {
	dispatch, err := sessionControl.QueryInterface(wca.IID_IAudioSessionControl2)
	if err != nil {
		return Session{}, false, fmt.Errorf("cannot get audio session control %d of device %v: %w", sessionIndex, this, err)
	}
	sessionControl2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))
	defer sessionControl2.Release()

	if err := sessionControl2.IsSystemSoundsSession(); err == nil {
		return Session{}, false, nil
	} else if oe, ok := common.AsError[*ole.OleError](err); ok && oe.Code() == uintptr(windows.ERROR_INVALID_FUNCTION) {
		// Ok, continue....
	} else {
		return Session{}, false, fmt.Errorf("cannot get determine if audio session %d of device %v is a system session or not: %w", sessionIndex, this, err)
	}

	var state uint32
	if err := sessionControl.GetState(&state); err != nil {
		return Session{}, false, fmt.Errorf("cannot get state of audio session %d of device %v: %w", sessionIndex, this, err)
	}

	switch state {
	case 1:
		var s Session
		if err := sessionControl2.GetProcessId(&s.HolderPid); err != nil {
			return Session{}, false, fmt.Errorf("cannot get PID of processes which hold session %d of device %v: %w", sessionIndex, this, err)
		}
		if err := sessionControl2.GetSessionIdentifier(&s.Identifier); err != nil {
			return Session{}, false, fmt.Errorf("cannot get session identifier of audio session %d of device %v: %w", sessionIndex, this, err)
		}

		if p, err := process.NewProcess(int32(s.HolderPid)); err == nil {
			if n, err := p.Name(); err == nil {
				s.Title = strings.TrimSuffix(n, ".exe")
			}
		}

		return s, true, nil
	default:
		return Session{}, false, nil
	}
}
