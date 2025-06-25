package common

import (
	"fmt"
	"regexp"
)

func NewRegexp(plain string) (result Regexp, err error) {
	err = result.Set(plain)
	return result, err
}

func MustNewRegexp(plain string) Regexp {
	result, err := NewRegexp(plain)
	if err != nil {
		panic(err)
	}
	return result
}

type Regexp struct {
	v *regexp.Regexp
}

func (this *Regexp) Set(plain string) error {
	if plain == "" {
		*this = Regexp{nil}
		return nil
	}

	buf, err := regexp.Compile(plain)
	if err != nil {
		return fmt.Errorf("illegal-regexp: %s", plain)
	}

	*this = Regexp{buf}
	return nil
}

func (this Regexp) String() string {
	if v := this.v; v != nil {
		return v.String()
	}
	return ""
}

func (this Regexp) MatchString(s string) bool {
	if v := this.v; v != nil {
		return v.MatchString(s)
	}
	return s == ""
}

func (this Regexp) MarshalText() (text []byte, err error) {
	return []byte(this.String()), nil
}

func (this *Regexp) UnmarshalText(text []byte) error {
	return this.Set(string(text))
}

func (this Regexp) IsZero() bool {
	return this.v == nil
}

func (this Regexp) HasContent() bool {
	return !this.IsZero()
}
