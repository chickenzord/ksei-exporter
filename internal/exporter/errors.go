package exporter

import (
	"fmt"
	"strings"
)

type Error struct {
	Account string
	Method  string
	Params  []string
	Cause   error
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s(%s) %s] %s", e.Method, strings.Join(e.Params, ", "), e.Account, e.Cause.Error())
}
