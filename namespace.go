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
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// ErrNoNamespace indicates that no namespace was provided.
var ErrNoNamespace = errors.New("no namespace provided")

type NamespaceError struct {
	Ns    []string
	ObjNs string
}

func (ns NamespaceError) Error() string {
	return fmt.Sprintf("Name '%s' not found in object (namespace=%s)", ns.ObjNs, strings.Join(ns.Ns, "."))
}

func IsNamespaceError(err error) bool {
	_, ok := err.(NamespaceError)
	return ok
}

type Stringer interface {
	String() string
}

// A value is a wrapper around a reflect value to provide panic safe methods.
type Value struct {
	reflect.Value
}

func (v *Value) CanFloat() bool {
	switch v.Value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// Float returns a float value is possible.
// Does kind checking to ensure a panic is avoided.
func (v *Value) Float() (float64, error) {
	switch v.Value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Value.Int()), nil
	case reflect.Float32, reflect.Float64:
		return v.Value.Float(), nil
	}
	return 0, errors.Errorf("kind %s is not a number", v.Value.Kind())
}

func (v *Value) String() string {
	if str, ok := v.Interface().(Stringer); ok {
		return str.String()
	}
	switch v.Kind() {
	case reflect.String:
		return v.Value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Interface())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Interface())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Interface())
	}
	return fmt.Sprintf("%s", v.Interface())
}

// StringNameSpace gets a value using the given namespace with a full-stop delimiter.
func StringNameSpace(i interface{}, namespace string) (*Value, error) {
	namespaces := strings.Split(namespace, ".")
	return NameSpace(i, namespaces...)
}

// NameSpace gets a value by the given namespaces in order.
func NameSpace(i interface{}, namespaces ...string) (*Value, error) {
	if len(namespaces) == 0 {
		return nil, ErrNoNamespace
	}
	v := reflect.ValueOf(i)
	for i := 0; i < len(namespaces); i++ {
		v = namespace(v, namespaces[i])
		if !v.IsValid() {
			return nil, NamespaceError{ObjNs: namespaces[i], Ns: namespaces}
		}
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
	}
	return &Value{Value: v}, nil
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
