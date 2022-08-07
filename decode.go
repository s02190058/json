package json

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Decode(input string, v interface{}) error {
	d, err := NewParser(input).Parse()
	if err != nil {
		return err
	}

	return d.Decode(v)
}

type Decoder struct {
	element
}

func addDecoderPrefixToError(err error) error {
	if err != nil {
		err = fmt.Errorf("decoder: %s", err.Error())
	}

	return err
}

func (d Decoder) Decode(v interface{}) error {
	return addDecoderPrefixToError(d.decode(v))
}

func invalidValue(t reflect.Type) error {
	if t == nil {
		return errors.New("unexpected nil")
	}

	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("unexpected non-pointer %s", t.String())
	}

	return fmt.Errorf("unexpected nil %s", t.String())
}

func (d Decoder) decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return invalidValue(reflect.TypeOf(v))
	}

	return d.element.toGo(rv.Elem())
}

func (str tString) toGo(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		t := v.Type().Elem()
		// checking for an available type to avoid unnecessary allocations
		switch t.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.String:
			if v.IsNil() {
				v.Set(reflect.New(t))
			}

			return str.toGo(v.Elem())
		}

		return fmt.Errorf("cannot decode JSON string into Go %s", t)

	case reflect.String:
		v.SetString(string(str))
		return nil

	case reflect.Interface:
		vv := reflect.New(reflect.TypeOf(string(str))).Elem()
		if err := str.toGo(vv); err != nil {
			return err
		}

		v.Set(vv)

		return nil
	}

	return fmt.Errorf("cannot decode JSON string into Go %s", v.Type())
}

func (num tNumber) toGo(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		t := v.Type().Elem()
		// checking for an available type to avoid unnecessary allocations
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64,
			reflect.Interface,
			reflect.Ptr:

			if v.IsNil() {
				v.Set(reflect.New(t))
			}

			return num.toGo(v.Elem())
		}

		return fmt.Errorf("cannot decode JSON number into Go %s", t)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(num))
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v.SetUint(uint64(num))
		return nil

	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(num))
		return nil

	case reflect.Interface:
		vv := reflect.New(reflect.TypeOf(float64(num))).Elem()
		if err := num.toGo(vv); err != nil {
			return err
		}

		v.Set(vv)

		return nil
	}

	return fmt.Errorf("cannot decode JSON number into Go %s", v.Type())
}

func (obj tObject) toGo(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		t := v.Type().Elem()
		// checking for an available type to avoid unnecessary allocations
		switch t.Kind() {
		case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Struct:
			if v.IsNil() {
				v.Set(reflect.New(t))
			}

			return obj.toGo(v.Elem())
		}

		return fmt.Errorf("cannot decode JSON object into Go %s", t)

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			tag := field.Tag
			data := tag.Get("json")
			name := strings.Split(data, ",")[0]

			if name == "-" {
				continue
			}

			if name == "" {
				name = strings.ToLower(field.Name)
			}

			value, ok := obj[name]
			if !ok {
				continue
			}

			if err := value.toGo(v.Field(i)); err != nil {
				return err
			}
		}

		return nil

	case reflect.Map:
		tk := v.Type().Key()
		switch tk.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.String:

		default:
			return fmt.Errorf("cannot decode JSON string into Go %s", tk)
		}

		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}

		for key, value := range obj {
			vk := reflect.New(tk).Elem()
			switch tk.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n, err := strconv.ParseInt(key, 10, 64)
				if err != nil {
					return fmt.Errorf("cannot decode JSON string into Go %s", tk)
				}

				vk.SetInt(n)

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				n, err := strconv.ParseUint(key, 10, 64)
				if err != nil {
					return fmt.Errorf("cannot decode JSON string into Go %s", tk)
				}

				vk.SetUint(n)

			case reflect.String:
				vk.SetString(key)

			default:
				return fmt.Errorf("cannot decode JSON string into Go %s", tk)
			}

			vv := reflect.New(v.Type().Elem()).Elem()
			if err := value.toGo(vv); err != nil {
				return err
			}

			v.SetMapIndex(vk, vv)
		}

		return nil

	case reflect.Interface:
		vv := reflect.New(reflect.TypeOf(map[string]interface{}{})).Elem()
		if err := obj.toGo(vv); err != nil {
			return err
		}

		v.Set(vv)

		return nil
	}

	return fmt.Errorf("cannot decode JSON object into Go %s", v.Type())
}

func (arr tArray) toGo(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		t := v.Type().Elem()
		// checking for an available type to avoid unnecessary allocations
		switch t.Kind() {
		case reflect.Array, reflect.Interface, reflect.Ptr, reflect.Slice:
			if v.IsNil() {
				v.Set(reflect.New(t))
			}

			return arr.toGo(v.Elem())
		}

		return fmt.Errorf("cannot decode JSON array into Go %s", t)

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if i >= len(arr) {
				break
			}

			value := arr[i]
			if err := value.toGo(v.Index(i)); err != nil {
				return err
			}
		}

		return nil

	case reflect.Slice:
		v.Set(v.Slice(0, 0))

		for _, value := range arr {
			vv := reflect.New(v.Type().Elem()).Elem()
			if err := value.toGo(vv); err != nil {
				return err
			}

			v.Set(reflect.Append(v, vv))
		}

		return nil

	case reflect.Interface:
		vv := reflect.New(reflect.TypeOf([]interface{}{})).Elem()
		if err := arr.toGo(vv); err != nil {
			return err
		}

		v.Set(vv)

		return nil
	}

	return fmt.Errorf("cannot decode JSON array into Go %s", v.Type())
}

func (boolean tBoolean) toGo(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		t := v.Type().Elem()
		// checking for an available type to avoid unnecessary allocations
		switch t.Kind() {
		case reflect.Bool, reflect.Interface, reflect.Ptr:
			if v.IsNil() {
				v.Set(reflect.New(t))
			}

			return boolean.toGo(v.Elem())
		}

		return fmt.Errorf("cannot decode JSON boolean into Go %s", t)

	case reflect.Bool:
		v.SetBool(bool(boolean))
		return nil

	case reflect.Interface:
		vv := reflect.New(reflect.TypeOf(bool(boolean))).Elem()
		if err := boolean.toGo(vv); err != nil {
			return err
		}

		v.Set(vv)

		return nil
	}

	return fmt.Errorf("cannot decode JSON boolean into Go %s", v.Type())
}

func (null tNull) toGo(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		v.Set(reflect.Zero(v.Type()))
	}

	return nil
}
