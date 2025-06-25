package audio

import "strings"

import "iter"

type Session struct {
	Identifier string `json:"identifier,omitempty"`
	HolderPid  uint32 `json:"pid,omitempty"`
	Title      string `json:"title,omitempty"`
	Device     Device `json:"-"`
}

func (this Session) CloneBare() Session {
	return Session{
		strings.Clone(this.Title),
		this.HolderPid,
		strings.Clone(this.Title),
		this.Device.CloneBare(),
	}
}

type Sessions []Session

func CollectSessionsErr(i iter.Seq2[*Session, error]) (Sessions, error) {
	var result Sessions
	for v, err := range i {
		if err != nil {
			return nil, err
		}
		result = append(result, v.CloneBare())
	}
	return result, nil
}

func (this Sessions) IsZero() bool {
	return len(this) <= 0
}

func (this Sessions) HasContent() bool {
	return !this.IsZero()
}

func (this Sessions) RelevantSessions(predicate func(*Session) (bool, error)) iter.Seq2[*Session, error] {
	return func(yield func(*Session, error) bool) {
		for _, candidate := range this {
			ok, err := predicate(&candidate)
			if err != nil {
				yield(nil, err)
				return
			}
			if ok {
				if !yield(&candidate, nil) {
					return
				}
			}
		}
	}
}

func (this Sessions) CloneBare() Sessions {
	result := make(Sessions, len(this))
	for i, v := range this {
		result[i] = v.CloneBare()
	}
	return result
}
