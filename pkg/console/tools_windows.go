package console

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	dllKernel32               = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleCtrlHandler = dllKernel32.NewProc("SetConsoleCtrlHandler")
	procExitThread            = dllKernel32.NewProc("ExitThread")
)

func SetConsoleCtrlHandler(h func(event any) bool) error {
	var r0 uintptr
	var err error
	if h == nil {
		r0, _, err = procSetConsoleCtrlHandler.Call(uintptr(0), uintptr(0))
	} else {
		hw := func(event uint32) uintptr {
			if !h(event) {
				return 1
			}
			switch event {
			case windows.CTRL_CLOSE_EVENT:
				exitThread(0)
				return 1
			}
			return 0
		}
		r0, _, err = procSetConsoleCtrlHandler.Call(syscall.NewCallback(hw), uintptr(1))
	}
	if r0 == 0 {
		return fmt.Errorf("cannot set console ctrl handler: %w", err)
	}
	return nil
}

func exitThread(exitCode uint32) {
	_, _, _ = procExitThread.Call(uintptr(exitCode))
}
