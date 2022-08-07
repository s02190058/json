package json

import (
	"reflect"
	"strings"
)

type element interface {
	toGo(v reflect.Value) error
	toJSON(b *strings.Builder)
}

type tString string

type tNumber float64

type tObject map[string]element

type tArray []element

type tBoolean bool

type tNull string
