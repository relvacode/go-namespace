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
	"github.com/renstrom/fuzzysearch/fuzzy"
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
	Suggestions []string
	Ns          string
}

func (ns NamespaceError) Error() string {
	s := fmt.Sprintf("Name %q not found in object", ns.Ns)
	if len(ns.Suggestions) > 0 {
		s = s + fmt.Sprintf(" (Did you mean %q?)", strings.Join(ns.Suggestions, ", "))
	}
	return s
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
		n := Get(v, namespaces[i])
		if !n.IsValid() {
			return Value{}, NamespaceError{Ns: namespaces[i], Suggestions: suggest(v, namespaces[i])}
		}
		v = n
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
	}
	return Value{Value: v}, nil
}

// suggest suggests the closest matches to the requested namespace name.
func suggest(v reflect.Value, name string) []string {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	targets := []string{}
	suggestions := []string{}
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				suggestions = append(suggestions, suggest(v.Field(i), name)...)
			}
			if strings.ToLower(f.Name) == strings.ToLower(name) {
				suggestions = append(suggestions, f.Name)
				continue
			}
			targets = append(targets, f.Name)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if k.Kind() != reflect.String {
				continue
			}
			targets = append(targets, k.String())
		}
	}
	return append(suggestions, fuzzy.Find(name, targets)...)
}

// Get gets a value from a given value using the given name.
func Get(v reflect.Value, name string) reflect.Value {
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
				nV := Get(v.Field(i), name)
				if nV.IsValid() {
					return nV
				}
			}
			ns := f.Tag.Get("ns")
			if ns == "-" {
				nV := Get(v.Field(i), name)
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
