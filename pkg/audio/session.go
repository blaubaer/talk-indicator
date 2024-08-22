package audio

type Session struct {
	Identifier string `json:"identifier,omitempty"`
	HolderPid  uint32 `json:"pid,omitempty"`
}

type Sessions []Session

func (this Sessions) IsZero() bool {
	return len(this) <= 0
}

func (this Sessions) HasContent() bool {
	return !this.IsZero()
}

func (this Sessions) HasRelevantSession(predicate func(*Session) bool) bool {
	hasAtLeastOneRelevantSession := false
	for _, session := range this {
		if predicate(&session) {
			hasAtLeastOneRelevantSession = true
			break
		}
	}
	return hasAtLeastOneRelevantSession
}
