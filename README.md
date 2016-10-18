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

// Get a value by a full stop delimited string value.
v, err := namespace.StringNameSpace(l, "PrimaryKey.SecondaryKey")
// Or get the value by a slice of namespaces.
v, err = namespace.NameSpace(l, "PrimaryKey", "SecondaryKey")

if err != nil {
  // Do something with the error here
}

fmt.Println(v.String()) // Outputs: MyValue
```
