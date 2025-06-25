package common

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

var (
	ErrLineTooLong   = errors.New("line too long")
	ErrStopIteration = errors.New("stop iteration")
)

func NewRingLineBuffer(maxLines, maxLineLength uint32) *RingLineBuffer {
	return &RingLineBuffer{
		current:         make([]byte, maxLineLength),
		currentCapacity: int(maxLineLength),
		lines:           make([][]byte, maxLines),
		linesCapacity:   int(maxLines),
	}
}

type RingLineBuffer struct {
	TruncateTooLongLines bool
	OnNewLine            func([]byte) ([]byte, error)

	currentLength   int
	current         []byte
	currentCapacity int

	linesOffset   int
	linesLength   int
	lines         [][]byte
	linesCapacity int

	mutex sync.RWMutex
}

func (this *RingLineBuffer) Write(p []byte) (n int, err error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for len(p) > 0 {
		var hasNl, treatAsNlAnyway bool
		rl := bytes.IndexRune(p, '\n')
		if rl < 0 {
			rl = len(p)
			hasNl = false
		} else {
			hasNl = true
		}

		if rl+this.currentLength > this.currentCapacity {
			if !this.TruncateTooLongLines {
				return n, ErrLineTooLong
			}
			// Truncate it...
			rl -= (rl + this.currentLength) - this.currentCapacity
			hasNl = false // As we truncate for at least one byte \n, cannot be longer a part of it.
			treatAsNlAnyway = true
		}
		copy(this.current[this.currentLength:], p[:rl])
		n += rl
		this.currentLength += rl
		if hasNl || treatAsNlAnyway {
			if hasNl {
				// We stripped of the \n of recording, but add it to the recording...
				n++
			}
			if err := this.addLine(this.current[:this.currentLength]); err != nil {
				return n, err
			}
			this.currentLength = 0
		}

		if hasNl {
			rl++
		}
		if len(p) <= rl {
			break
		}
		p = p[rl:]
	}

	return n, nil
}

func (this *RingLineBuffer) AddLine(line []byte) error {
	if len(line) > this.currentCapacity {
		return ErrLineTooLong
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.addLine(line)
}

func (this *RingLineBuffer) addLine(line []byte) error {
	if v := this.OnNewLine; v != nil {
		var err error
		if line, err = v(line); err != nil {
			return err
		}
	}

	if this.linesLength >= this.linesCapacity {
		this.linesOffset++
		if this.linesOffset > this.linesCapacity {
			this.linesOffset -= this.linesCapacity
		}
	}
	i := this.linesOffset + this.linesLength
	if i >= this.linesCapacity {
		i -= this.linesCapacity + 1
	}

	this.lines[i] = bytes.Clone(line)
	if this.linesLength < this.linesCapacity {
		this.linesLength++
	}

	return nil
}

func (this *RingLineBuffer) NumberOfLines() uint32 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.numberOfLines()
}

func (this *RingLineBuffer) numberOfLines() uint32 {
	return uint32(this.linesLength)
}

type LineConsumer func(uint32, []byte) error

func (this *RingLineBuffer) WriteTo(to io.Writer) (n int64, err error) {
	err = this.ConsumeLines(func(_ uint32, line []byte) error {
		wn, wErr := to.Write(append(line, '\n'))
		n += int64(wn)
		return wErr
	})

	return n, nil
}

func (this *RingLineBuffer) ConsumeLines(consumer LineConsumer) error {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.consumeLines(consumer)
}

func (this *RingLineBuffer) consumeLines(consumer LineConsumer) error {
	consume := func(i int, line []byte) (bool, error) {
		if err := consumer(uint32(i), line); err != nil {
			if errors.Is(err, ErrStopIteration) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}

	end := this.linesOffset + this.linesLength
	if this.linesCapacity > end {
		for i, line := range this.lines[this.linesOffset:end] {
			if canContinue, err := consume(i, line); err != nil || !canContinue {
				return err
			}
		}
		return nil
	}

	var i int
	for _, line := range this.lines[this.linesOffset:] {
		if canContinue, err := consume(i, line); err != nil || !canContinue {
			return err
		}
		i++
	}
	for _, line := range this.lines[0:end] {
		if canContinue, err := consume(i, line); err != nil || !canContinue {
			return err
		}
		i++
	}

	return nil
}
