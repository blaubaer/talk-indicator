package signal

import (
	"github.com/blaubaer/talk-indicator/pkg/audio"
	"iter"
)

type Context interface {
	State() State
	Sessions() iter.Seq2[*audio.Session, error]
}
