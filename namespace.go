// Package namespace is a utility to retrieve a Go value from a namespace value.
// Note that due to the laws of reflection only public fields can be accessed by namespace.
// Struct fields can use the tag ns to modify how namespace handles that struct field.
// - indicates that the namespace should pass through, treating the field as a transparent name.
//
//    struct {
//      Sub SubType `ns:"-"
//    }
//
// Any other value renames the namespace name of that field.
//
//    struct {
//      Sub SubType `ns:"something"
//    }
//
package namespace

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Kinder is an interface that reports its kind.
type Kinder interface {
	Kind() reflect.Kind
}

// ErrNoNamespace indicates that no namespace was provided.
var ErrNoNamespace = errors.New("no namespace provided")

type NamespaceError struct {
	Ns    []string
	ObjNs string
}

func (ns NamespaceError) Error() string {
	return fmt.Sprintf("Name '%s' not found in object (namespace=%s)", ns.ObjNs, strings.Join(ns.Ns, "."))
}

func ValueOf(v interface{}) Value {
	return Value{Value: reflect.ValueOf(v)}
}

// IsNumber returns true if the given reflect value is or can be converted to a number.
func IsNumber(v Kinder) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// A value is a wrapper around a reflect value to provide panic safe methods.
type Value struct {
	reflect.Value
}

// Float returns a float value if possible.
func (v Value) Float() (float64, error) {
	switch v.Value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Value.Int()), nil
	case reflect.Float32, reflect.Float64:
		return v.Value.Float(), nil
	}
	return 0, fmt.Errorf("Kind %s is not a float", v.Value.Kind())
}

// Int returns an int value if possible
func (v Value) Int() (int64, error) {
	switch v.Value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Value.Int(), nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Value.Float()), nil
	}
	return 0, fmt.Errorf("Kind %s is not an int", v.Value.Kind())
}

// String uses fmt.Sprint to return a string representation of the value.
// Unless the kind is already a string in which that is used instead.
func (v Value) String() string {
	if v.Kind() == reflect.String {
		return v.Value.String()
	}
	return fmt.Sprint(v.Interface())
}

// Namespace gets a value by the given namespaces in order.
func Namespace(i interface{}, namespaces []string) (Value, error) {
	if len(namespaces) == 0 {
		return Value{}, ErrNoNamespace
	}
	v := reflect.ValueOf(i)
	for i := 0; i < len(namespaces); i++ {
		v = namespace(v, namespaces[i])
		if !v.IsValid() {
			return Value{}, NamespaceError{ObjNs: namespaces[i], Ns: namespaces}
		}
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
	}
	return Value{Value: v}, nil
}

func namespace(v reflect.Value, name string) reflect.Value {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// dereference pointers
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		typ := v.Type()
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Anonymous {
				nV := namespace(v.Field(i), name)
				if nV.IsValid() {
					return nV
				}
			}
			ns := f.Tag.Get("ns")
			if ns == "-" {
				nV := namespace(v.Field(i), name)
				if nV.IsValid() {
					return nV
				}
			}
			if ns != "" && name == ns {
				return v.Field(i)
			}
			if f.Name == name {
				return v.Field(i)
			}
		}
	case reflect.Map:
		return v.MapIndex(reflect.ValueOf(name))
	}
	return reflect.Value{}
}
