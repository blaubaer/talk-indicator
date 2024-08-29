package audio

import "strings"

type Session struct {
	Identifier string `json:"identifier,omitempty"`
	HolderPid  uint32 `json:"pid,omitempty"`
	Title      string `json:"title,omitempty"`
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

func (this Sessions) Filter(predicate func(*Session) bool) Sessions {
	var result Sessions
	for _, session := range this {
		if predicate(&session) {
			result = append(result, Session{
				strings.Clone(session.Title),
				session.HolderPid,
				strings.Clone(session.Title),
			})
		}
	}
	return result
}
