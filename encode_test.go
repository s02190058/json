package json

import (
	"testing"
)

func TestEncode(t *testing.T) {
	type SimpleStruct struct {
		String string
		Int    int
		Float  float64
		Bool   bool
		IntPtr *int
	}

	simpleStruct := SimpleStruct{
		String: "simple",
		Int:    100500,
		Float:  3.14,
		Bool:   true,
		IntPtr: nil,
	}

	simpleMap := map[string]interface{}{
		"string": "simple",
		"int":    100500,
		"float":  3.14,
		"bool":   true,
		"intptr": nil,
	}

	structMapRes := `{"bool":true,"float":3.14,"int":100500,"intptr":null,"string":"simple"}`

	simpleArray := [5]interface{}{"simple", 100500, 3.14, true, nil}

	simpleSlice := []interface{}{"simple", 100500, 3.14, true, nil}

	arraySliceRes := `["simple",100500,3.14,true,null]`

	testCases := []struct {
		name    string
		v       interface{}
		isValid bool
		res     string
	}{
		{
			name:    "string",
			v:       "string",
			isValid: true,
			res:     `"string"`,
		},
		{
			name:    "int",
			v:       100500,
			isValid: true,
			res:     "100500",
		},
		{
			name:    "float",
			v:       3.14,
			isValid: true,
			res:     "3.14",
		},
		{
			name:    "float with insignificant zeros",
			v:       3.14000,
			isValid: true,
			res:     "3.14",
		},
		{
			name:    "bool: true",
			v:       true,
			isValid: true,
			res:     "true",
		},
		{
			name:    "bool: false",
			v:       false,
			isValid: true,
			res:     "false",
		},
		{
			name:    "simple struct",
			v:       simpleStruct,
			isValid: true,
			res:     structMapRes,
		},
		{
			name:    "simple map",
			v:       simpleMap,
			isValid: true,
			res:     structMapRes,
		},
		{
			name: "complex struct",
			v: struct {
				IntPtr       *int
				Interface    interface{}
				SimpleStruct *SimpleStruct
			}{
				IntPtr:       new(int),
				Interface:    &simpleMap,
				SimpleStruct: &simpleStruct,
			},
			isValid: true,
			res:     `{"interface":` + structMapRes + `,"intptr":0,"simplestruct":` + structMapRes + "}",
		},
		{
			name: "complex map",
			v: map[string]interface{}{
				"intptr":       new(int),
				"interface":    &simpleMap,
				"simplestruct": &simpleStruct,
			},
			isValid: true,
			res:     `{"interface":` + structMapRes + `,"intptr":0,"simplestruct":` + structMapRes + "}",
		},
		{
			name:    "simple array",
			v:       simpleArray,
			isValid: true,
			res:     arraySliceRes,
		},
		{
			name:    "simple slice",
			v:       simpleSlice,
			isValid: true,
			res:     arraySliceRes,
		},
		{
			name: "complex array",
			v: [4]interface{}{
				&simpleStruct,
				&simpleMap,
				&simpleArray,
				&simpleSlice,
			},
			isValid: true,
			res:     "[" + structMapRes + "," + structMapRes + "," + arraySliceRes + "," + arraySliceRes + "]",
		},
		{
			name: "complex slice",
			v: []interface{}{
				&simpleStruct,
				&simpleMap,
				&simpleArray,
				&simpleSlice,
			},
			isValid: true,
			res:     "[" + structMapRes + "," + structMapRes + "," + arraySliceRes + "," + arraySliceRes + "]",
		},
		{
			name:    "invalid type",
			v:       make(chan int),
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Encode(tc.v)
			if tc.isValid {
				if err != nil {
					t.Error(err)

				}
				if res != tc.res {
					t.Errorf("Encode(v) FAILED. Expected:\n%s\ngot:\n%s", tc.res, res)
				}
			} else {
				if err == nil {
					t.Errorf("Encode(v) FAILED. Expected not nil error")
				}
			}

		})
	}
}
