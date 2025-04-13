package utils

import (
	"fmt"
	"reflect"
	"strings"
)

func ValidateFields(input any) string {
	var invalidProps []string

	// Get the reflect.Value of the input and dereference if it's a pointer.
	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Ensure we are dealing with a struct.
	if v.Kind() != reflect.Struct {
		return "input is not a struct"
	}

	// Loop through each field in the struct.
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		// Only check exported fields.
		if !fieldValue.CanInterface() {
			continue
		}

		// Check for blank strings.
		if fieldValue.Kind() == reflect.String && fieldValue.String() == "" {
			invalidProps = append(invalidProps, fmt.Sprintf("%s is blank", fieldType.Name))
		}

		// Check for nil pointers or interfaces.
		if (fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Interface) && fieldValue.IsNil() {
			invalidProps = append(invalidProps, fmt.Sprintf("%s is nil", fieldType.Name))
		}

		// You can add more checks here for other types as needed.
	}

	if len(invalidProps) > 0 {
		return strings.Join(invalidProps, ", ")
	}
	return ""
}
