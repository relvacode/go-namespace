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
	"github.com/renstrom/fuzzysearch/fuzzy"
	"reflect"
	"strings"
)

// Kinder is an interface that reports its kind.
type Kinder interface {
	Kind() reflect.Kind
}

// A Namespacer is an object that can retrieve it's own namespace value.
// If a type implements Namespacer then that method is used instead of reflect traversal.
type Namespacer interface {
	Namespace([]string) (Value, error)
}

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

// ValueOf creates a new Value from the given interface
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
// If the length of namespace is emtpy then the object itself is returned.
func Namespace(i interface{}, namespaces []string) (Value, error) {
	if ns, ok := i.(Namespacer); ok {
		return ns.Namespace(namespaces)
	}
	v := reflect.ValueOf(i)
	for i := 0; i < len(namespaces); i++ {
		if ns, ok := v.Interface().(Namespacer); ok {
			return ns.Namespace(namespaces[i:])
		}
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

// Field returns the namespace name for a given struct field.
func Field(v reflect.StructField) string {
	ns := v.Tag.Get("ns")
	if ns != "" {
		return ns
	}
	return v.Name
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
			ns := Field(f)
			if (f.Anonymous && ns != "") || ns == "-" {
				nV := Get(v.Field(i), name)
				if nV.IsValid() {
					return nV
				}
			}
			if ns == name {
				return v.Field(i)
			}
		}
	case reflect.Map:
		return v.MapIndex(reflect.ValueOf(name))
	}
	return reflect.Value{}
}

// Names returns a list of all possible namespaces for the given object.
// nil is returned if the object does not contain namespaces.
func Names(v interface{}) [][]string {
	return names(reflect.TypeOf(v), nil)
}

func names(v reflect.Type, prev []string) (ns [][]string) {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			n := Field(f)
			if f.Anonymous && n != "" || n == "-" {
				ns = append(ns, names(f.Type, prev)...)
				continue
			}
			tn := names(f.Type, append(prev, n))
			if tn == nil {
				ns = append(ns, append(prev, n))
				continue
			}
			ns = append(ns, tn...)
		}
	}
	return
}
