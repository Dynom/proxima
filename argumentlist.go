package main

import "strings"

type argumentList []string

func (l argumentList) String() string {
	return strings.Join(l, ",")
}

func (a *argumentList) Set(value string) error {
	*a = append(*a, strings.Split(value, ",")...)
	return nil
}
