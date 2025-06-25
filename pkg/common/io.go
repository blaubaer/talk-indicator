package common

import (
	"fmt"
	"github.com/chzyer/readline"
	log "github.com/echocat/slf4g"
	"os"
)

type settable interface {
	IsZero() bool
	Set(string) error
}

func RequestContentIfRequiredFromTerminal(of settable, promptName string, canBeEmpty, isPassword bool) error {
	if of.IsZero() {
		l, err := readline.NewEx(&readline.Config{
			Stdin:  os.Stdin,
			Stdout: os.Stderr,
		})
		if err != nil {
			return fmt.Errorf("could not read from terminal for prompt %q: %w", promptName, err)
		}
		defer func() {
			_ = l.Close()
		}()

		prompt := fmt.Sprintf("Enter %s: ", promptName)
		l.SetPrompt(prompt)
		if isPassword {
			l.SetMaskRune('*')
		}
		l.ResetHistory()
		for of.IsZero() {
			var line string
			if isPassword {
				var b []byte
				b, err = l.ReadPassword(prompt)
				line = string(b)
			} else {
				line, err = l.Readline()
			}
			if err != nil {
				return fmt.Errorf("could not read from terminal for prompt %q: %w", promptName, err)
			}
			if err := of.Set(line); err != nil {
				log.WithError(err).
					Error()
			}
			if canBeEmpty && of.IsZero() {
				return nil
			}
		}
	}
	return nil
}

func RequestStringContentIfRequiredFromTerminal(of *string, promptName string, canBeEmpty, isPassword bool) error {
	buf := rawString(*of)
	if err := RequestContentIfRequiredFromTerminal(&buf, promptName, canBeEmpty, isPassword); err != nil {
		return err
	}
	*of = string(buf)
	return nil
}

func RequestRawStringContentIfRequiredFromTerminal(of *[]byte, promptName string, canBeEmpty, isPassword bool) error {
	buf := rawString(*of)
	if err := RequestContentIfRequiredFromTerminal(&buf, promptName, canBeEmpty, isPassword); err != nil {
		return err
	}
	*of = buf
	return nil
}

type rawString []byte

func (v rawString) IsZero() bool {
	return len(v) == 0
}

func (v *rawString) Set(s string) error {
	*v = rawString(s)
	return nil
}
