package main

import "strings"

type argumentList []string

func (al argumentList) String() string {
	return strings.Join(al, ",")
}

func (al argumentList) PrettyString() string {
	if len(al) == 0 {
		return "*"
	}

	return strings.Join(al, ",")
}

func (al *argumentList) Set(value string) error {
	*al = append(*al, strings.Split(value, ",")...)
	return nil
}
