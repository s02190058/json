package json

import (
	"sort"
	"strconv"
	"strings"
)

type Composer struct {
	element
}

func (c *Composer) Compose() string {
	return c.compose()
}

func (c *Composer) compose() string {
	b := &strings.Builder{}
	c.element.toJSON(b)

	return b.String()
}

func (str tString) toJSON(b *strings.Builder) {
	b.WriteString("\"" + string(str) + "\"")
}

func (num tNumber) toJSON(b *strings.Builder) {
	b.WriteString(strconv.FormatFloat(float64(num), 'f', -1, 64))
}

func (obj tObject) toJSON(b *strings.Builder) {
	b.WriteByte('{')

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	i := 0
	for _, key := range keys {
		b.WriteString("\"" + key + "\"")
		b.WriteByte(':')
		obj[key].toJSON(b)

		if i < len(obj)-1 {
			b.WriteByte(',')
		}
		i++
	}

	b.WriteByte('}')
}

func (arr tArray) toJSON(b *strings.Builder) {
	b.WriteByte('[')

	for i, value := range arr {
		value.toJSON(b)

		if i < len(arr)-1 {
			b.WriteByte(',')
		}
	}

	b.WriteByte(']')
}

func (boolean tBoolean) toJSON(b *strings.Builder) {
	b.WriteString(strconv.FormatBool(bool(boolean)))
}

func (null tNull) toJSON(b *strings.Builder) {
	b.WriteString("null")
}
