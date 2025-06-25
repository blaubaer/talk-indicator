package console

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ErrAllocationFailed = errors.New("allocation failed")

	procAllocConsole    = dllKernel32.NewProc("AllocConsole")
	procSetConsoleTitle = dllKernel32.NewProc("SetConsoleTitleW")
	procFreeConsole     = dllKernel32.NewProc("FreeConsole")
)

func NewDedicatedConsole(title string) (*DedicatedConsole, error) {
	if r0, _, err := procAllocConsole.Call(); r0 == 0 {
		return nil, fmt.Errorf("%w: %v", ErrAllocationFailed, err)
	}

	titlePtr, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return nil, fmt.Errorf("cannot allocate title: %w", err)
	}

	if r0, _, err := procSetConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr))); r0 == 0 {
		return nil, fmt.Errorf("cannot set console title: %w", err)
	}

	hIn, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return nil, fmt.Errorf("cannot get stdin handle: %w", err)
	}
	hOut, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return nil, fmt.Errorf("cannot get stdout handle: %w", err)
	}
	hErr, err := windows.GetStdHandle(windows.STD_ERROR_HANDLE)
	if err != nil {
		return nil, fmt.Errorf("cannot get stderr handle: %w", err)
	}

	var inT, outT, errT uint32
	if inT, err = configureHandle(hIn, windows.STD_INPUT_HANDLE); err != nil {
		return nil, fmt.Errorf("cannot enable virtual terminal processing on stdin: %w", err)
	}
	if outT, err = configureHandle(hOut, windows.STD_OUTPUT_HANDLE); err != nil {
		return nil, fmt.Errorf("cannot enable virtual terminal processing on stdout: %w", err)
	}
	if errT, err = configureHandle(hErr, windows.STD_ERROR_HANDLE); err != nil {
		return nil, fmt.Errorf("cannot enable virtual terminal processing on stderr: %w", err)
	}

	result := &DedicatedConsole{
		Stdin:  os.NewFile(uintptr(hIn), "/dev/stdin"),
		Stdout: os.NewFile(uintptr(hOut), "/dev/stdout"),
		Stderr: os.NewFile(uintptr(hErr), "/dev/stderr"),

		StdinMode:  inT,
		StdoutMode: outT,
		StderrMode: errT,
	}

	if err := SetConsoleCtrlHandler(result.onCtrlC); err != nil {
		return nil, fmt.Errorf("cannot set console ctrl handler: %w", err)
	}

	return result, nil
}

func configureHandle(handle windows.Handle, ht windows.Handle) (mode uint32, _ error) {
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return 0, nil
	}

	tMode := mode
	switch ht {
	case windows.STD_INPUT_HANDLE:
		tMode = windows.ENABLE_VIRTUAL_TERMINAL_INPUT |
			windows.ENABLE_PROCESSED_INPUT |
			windows.ENABLE_MOUSE_INPUT |
			windows.ENABLE_EXTENDED_FLAGS
	case windows.STD_OUTPUT_HANDLE, windows.STD_ERROR_HANDLE:
		tMode = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING |
			windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING |
			windows.ENABLE_WRAP_AT_EOL_OUTPUT |
			windows.ENABLE_PROCESSED_OUTPUT
	default:
		panic(fmt.Errorf("don't know how to handle %d", handle))
	}

	if err := windows.SetConsoleMode(handle, tMode); err != nil {
		return mode, err
	}

	return mode, nil
}

func (this *DedicatedConsole) Close() (err error) {
	c := func(what io.Closer) {
		if cErr := what.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}
	defer func() {
		defer func() {
			if r := recover(); r != nil {
				if eErr, ok := r.(error); ok {
					err = eErr
				} else {
					err = fmt.Errorf("%v", r)
				}
			}
		}()
		_, _, _ = procFreeConsole.Call()
	}()
	defer func() {
		_ = SetConsoleCtrlHandler(nil)
	}()
	defer c(this.Stdin)
	defer c(this.Stdout)
	defer c(this.Stderr)

	return nil
}
