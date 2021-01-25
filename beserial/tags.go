package beserial

import (
	"reflect"
	"strings"
)

type tags struct {
	tagName  string
	lenTag   reflect.Kind
	optional bool
}

func (ts *tags) parse(tag string) {
	const optional = "optional"
	const lenTag = "len_tag="

	elems := strings.Split(tag, ",")
	for _, el := range elems {
		switch {
		case el == optional:
			ts.optional = true
		case strings.HasPrefix(tag, lenTag):
			switch tag[len(lenTag):] {
			case "uint8":
				ts.lenTag = reflect.Uint8
			case "uint16":
				ts.lenTag = reflect.Uint16
			case "uint32":
				ts.lenTag = reflect.Uint32
			case "uint64":
				ts.lenTag = reflect.Uint64
			default:
				panic("beserial: invalid len_tag: " + lenTag)
			}
		default:
			panic("beserial: unknown struct tag option: " + el)
		}
	}
}
