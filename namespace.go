package namespace

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

var ErrNoNamespace = errors.New("no namespace provided")

type Value struct {
	reflect.Value
}

func (v *Value) Float() (float64, error) {
	switch v.Value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Value.Int()), nil
	case reflect.Float32, reflect.Float64:
		return v.Value.Float(), nil
	}
	return 0, errors.Errorf("kind %s is not a number", v.Value.Kind())
}

func GetString(i interface{}, namespace string) (*Value, error) {
	namespaces := strings.Split(namespace, ".")
	if len(namespaces) == 0 {
		return nil, ErrNoNamespace
	}
	return Get(i, namespaces...)
}

func namespace(v reflect.Value, name string) reflect.Value {
	// dererence interfaces
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

func Get(i interface{}, namespaces ...string) (*Value, error) {
	v := reflect.ValueOf(i)
	for i := 0; i < len(namespaces); i++ {
		v = namespace(v, namespaces[i])
		if !v.IsValid() {
			return nil, errors.Errorf("namespace %s not found", namespaces[i])
		}
		if v.Kind() == reflect.Interface {
			v = v.Elem()
		}
	}
	return &Value{Value: v}, nil
}
