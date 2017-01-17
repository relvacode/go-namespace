```go
import "github.com/relvacode/go-namespace"
```

[![GoDoc](https://godoc.org/github.com/relvacode/go-namespace?status.svg)](https://godoc.org/github.com/relvacode/go-namespace)
[![Build Status](https://travis-ci.org/relvacode/go-namespace.svg?branch=master)](https://travis-ci.org/relvacode/go-namespace)

Go namespace provides the ability to get a Go value by a string name. Similar to `text/template`'s value access without templating.

```go
// l is your value that you want to search its namespace for
// It can be either a struct or map, or an interface to a struct or map.
l := map[string]interface{}{
  "PrimaryKey": map[string]interface{}{
    "SecondaryKey": "MyValue",
  },
}

// Get a value using ordered names.
v, err = namespace.NameSpace(l, []string{"PrimaryKey", "SecondaryKey"})

if err != nil {
  // Do something with the error here
}

fmt.Println(v.String()) // Outputs: MyValue
```

### Struct Tags

Using the `ns` struct tag you can change how namespace will access that struct field.

```go
type MyStruct struct {
        // Ignore this field as a namespace
        Passthough OtherStruct `ns:'-"`
        
        // Rename the namespace name to 'new'
        Rename OtherStruct `ns:"new"`
}
```

### `namespace.Value`

`namespace.Value` is a wrapper around `reflect.Value` that provides safe methods for converting objects into a requested native type.

```go
v := namespace.ValueOf(1234.567)
i, err := v.Int()
```
