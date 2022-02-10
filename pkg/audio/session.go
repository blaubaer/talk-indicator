package audio

type Session struct {
	HolderPid uint32 `json:"pid,omitempty"`
}

type Sessions []Session

func (this Sessions) IsZero() bool {
	return len(this) <= 0
}

func (this Sessions) HasContent() bool {
	return !this.IsZero()
}
