package common

import "iter"

func Iter2Err[V any](in ...V) iter.Seq2[V, error] {
	return func(yield func(V, error) bool) {
		for _, v := range in {
			if !yield(v, nil) {
				return
			}
		}
	}
}

func Iter[V any](in ...V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range in {
			if !yield(v) {
				return
			}
		}
	}
}
