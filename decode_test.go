package json

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	simpleObject := `{"bool":true,"float":3.14,"int":100500,"intptr":null,"string":"simple"}`

	type SimpleStruct struct {
		String string
		Int    int
		Float  float64
		Bool   bool
		IntPtr *int
	}

	structRes := SimpleStruct{
		String: "simple",
		Int:    100500,
		Float:  3.14,
		Bool:   true,
		IntPtr: nil,
	}

	mapRes := map[string]interface{}{
		"string": "simple",
		"int":    float64(100500),
		"float":  3.14,
		"bool":   true,
		"intptr": nil,
	}

	simpleArray := `["simple",100500,3.14,true,null]`

	arrayRes := [5]interface{}{"simple", float64(100500), 3.14, true, nil}

	sliceRes := []interface{}{"simple", float64(100500), 3.14, true, nil}

	type ComplexStruct struct {
		Struct *SimpleStruct
		Map    *map[string]interface{}
	}

	testCases := []struct {
		name    string
		input   string
		v       interface{}
		isValid bool
		res     interface{}
	}{
		{
			name:    "simple object to struct",
			input:   simpleObject,
			v:       &SimpleStruct{},
			isValid: true,
			res:     &structRes,
		},
		{
			name:    "simple object to map",
			input:   simpleObject,
			v:       &map[string]interface{}{},
			isValid: true,
			res:     &mapRes,
		},
		{
			name:    "complex object to struct",
			input:   `{"struct":` + simpleObject + `,"map":` + simpleObject + "}",
			v:       &ComplexStruct{},
			isValid: true,
			res: &ComplexStruct{
				Struct: &structRes,
				Map:    &mapRes,
			},
		},
		{
			name:    "complex array to slice",
			input:   "[" + simpleObject + "," + simpleArray + "]",
			v:       &[]interface{}{},
			isValid: true,
			res: &[]interface{}{
				mapRes,
				sliceRes,
			},
		},
		{
			name:    "simple array to array",
			input:   simpleArray,
			v:       &[5]interface{}{},
			isValid: true,
			res:     &arrayRes,
		},
		{
			name:    "simple array to slice",
			input:   simpleArray,
			v:       &[]interface{}{},
			isValid: true,
			res:     &sliceRes,
		},
		{
			name:    "unknown identifier",
			input:   `{"key":value}`,
			v:       &struct{}{},
			isValid: false,
		},
		{
			name:    "object with lost colon",
			input:   `{"key1":"value1","key2" "value2"}`,
			v:       &struct{}{},
			isValid: false,
		},
		{
			name:    "object with lost comma",
			input:   `{"key1":"value1" "key2":"value2"}`,
			v:       &struct{}{},
			isValid: false,
		},
		{
			name:    "object with lost closing curly bracket",
			input:   `{"key1":"value1","key2":"value2"`,
			v:       &struct{}{},
			isValid: false,
		},
		{
			name:    "array with lost comma",
			input:   `["value1" "value2"]`,
			v:       &[]interface{}{},
			isValid: false,
		},
		{
			name:    "array with lost closing square bracket",
			input:   `["value1","value2"`,
			v:       &[]interface{}{},
			isValid: false,
		},
		{
			name:    "forget pointer",
			input:   simpleObject,
			v:       SimpleStruct{},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Decode(tc.input, tc.v)
			if tc.isValid {
				if err != nil {
					t.Error(err)

				}
				if !reflect.DeepEqual(tc.res, tc.v) {
					t.Errorf("Decode(input, v) FAILED. Expected:\n%s\ngot:\n%s", tc.res, tc.v)
				}
			} else {
				if err == nil {
					t.Errorf("Decode(input, v) FAILED. Expected not nil error")
				}
			}
		})
	}

}
