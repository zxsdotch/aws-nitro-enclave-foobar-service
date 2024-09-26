package utils

import "log"

func PanicOnErr(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func Ref[T any](v T) *T {
	return &v
}
