package common

import "github.com/alecthomas/kingpin/v2"

type FlagHolder interface {
	Flag(name, help string) *kingpin.FlagClause
}
