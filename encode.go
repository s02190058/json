package json

import (
	"fmt"
	"reflect"
	"strings"
)

func Encode(v interface{}) (string, error) {
	c, err := NewEncoder().Encode(v)
	if err != nil {
		return "", err
	}

	return c.Compose(), nil
}

type Encoder struct{}

func NewEncoder() Encoder {
	return Encoder{}
}

func addEncoderPrefixToError(err error) error {
	if err != nil {
		err = fmt.Errorf("encoder: %s", err.Error())
	}

	return err
}

func (e Encoder) Encode(v interface{}) (Composer, error) {
	c, err := e.encode(v)
	err = addEncoderPrefixToError(err)
	return Composer{c}, err
}

func (e Encoder) encode(v interface{}) (element, error) {
	rv := reflect.ValueOf(v)
	return e.encodePrimary(rv)
}

func (e Encoder) encodePrimary(v reflect.Value) (element, error) {
	return e.encodeValue(v)
}

func (e Encoder) encodeObject(v reflect.Value) (element, error) {
	switch v.Kind() {
	case reflect.Struct:
		pairs := make(map[string]element)
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			tag := field.Tag
			data := tag.Get("json")
			dataParts := strings.Split(data, ",")

			if len(dataParts) > 1 {
				if dataParts[1] == "omitempty" && v.IsZero() {
					continue
				}
			}

			name := dataParts[0]

			if name == "-" {
				continue
			}

			if name == "" {
				name = strings.ToLower(field.Name)
			}

			value, err := e.encodeValue(v.Field(i))
			if err != nil {
				return nil, err
			}

			pairs[name] = value
		}

		return tObject(pairs), nil

	case reflect.Map:
		pairs := make(map[string]element)
		for _, vk := range v.MapKeys() {
			var key string
			tk := vk.Type()
			switch tk.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				key = fmt.Sprintf("%d", vk.Int())

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				key = fmt.Sprintf("%d", vk.Uint())

			case reflect.String:
				key = vk.String()

			default:
				return nil, fmt.Errorf("unexpected %s", tk)
			}

			value, err := e.encodeValue(v.MapIndex(vk))
			if err != nil {
				return nil, err
			}

			pairs[key] = value
		}

		return tObject(pairs), nil
	}

	// in general, it should not happen
	return nil, fmt.Errorf("(internal) unexpected %s", v.Type())
}

func (e Encoder) encodeArray(v reflect.Value) (element, error) {
	switch v.Type().Kind() {
	case reflect.Array, reflect.Slice:
		values := make([]element, 0)
		for i := 0; i < v.Len(); i++ {
			value, err := e.encodeValue(v.Index(i))
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}

		return tArray(values), nil
	}

	// in general, it should not happen
	return nil, fmt.Errorf("(internal) unexpected %s", v.Type())
}

func (e Encoder) encodeValue(v reflect.Value) (element, error) {
	switch v.Kind() {
	case reflect.Invalid:
		return tNull("null"), nil

	case reflect.Bool:
		return tBoolean(v.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return tNumber(v.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return tNumber(v.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return tNumber(v.Float()), nil

	case reflect.Array:
		return e.encodeArray(v)

	case reflect.Interface:
		if v.IsNil() {
			return tNull("null"), nil
		}

		return e.encodeValue(v.Elem())

	case reflect.Map:
		if v.IsNil() {
			return tNull("null"), nil
		}

		return e.encodeObject(v)

	case reflect.Ptr:
		if v.IsNil() {
			return tNull("null"), nil
		}

		return e.encodeValue(v.Elem())

	case reflect.Slice:
		if v.IsNil() {
			return tNull("null"), nil
		}

		return e.encodeArray(v)

	case reflect.String:
		return tString(v.String()), nil

	case reflect.Struct:
		return e.encodeObject(v)
	}

	return nil, fmt.Errorf("unexpected %s", v.Type())
}
