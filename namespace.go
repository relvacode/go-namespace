// package namespace is a utility to retrieve a Go value from a string namespace.
package namespace

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

// ErrNoNamespace indicates that no namespace was provided.
var ErrNoNamespace = errors.New("no namespace provided")

// A value is a wrapper around a reflect value to provide panic safe methods.
type Value struct {
	reflect.Value
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

// Get a value using the given namespace with a full-stop delimiter.
func GetString(i interface{}, namespace string) (*Value, error) {
	namespaces := strings.Split(namespace, ".")
	return Get(i, namespaces...)
}

// Get a value by the given namespaces in order.
func Get(i interface{}, namespaces ...string) (*Value, error) {
	if len(namespaces) == 0 {
		return nil, ErrNoNamespace
	}
	v := reflect.ValueOf(i)
	for i := 0; i < len(namespaces); i++ {
		v = namespace(v, namespaces[i])
		if !v.IsValid() {
			return nil, errors.Errorf("namespace %s not found", strings.Join(namespaces[:i], "."))
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
		return v.FieldByName(name)
	case reflect.Map:
		return v.MapIndex(reflect.ValueOf(name))
	}
	return reflect.Value{}
}
