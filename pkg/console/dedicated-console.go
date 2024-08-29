package console

import (
	"os"
)

type DedicatedConsole struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File

	StdinMode  uint32
	StdoutMode uint32
	StderrMode uint32

	OnCtrlC func(event any) bool
}

func (this *DedicatedConsole) Read(p []byte) (n int, err error) {
	return this.Stdin.Read(p)
}

func (this *DedicatedConsole) Write(p []byte) (n int, err error) {
	return this.Stdout.Write(p)
}

func (this *DedicatedConsole) onCtrlC(event any) bool {
	if v := this.OnCtrlC; v != nil {
		return v(event)
	}
	return true
}
