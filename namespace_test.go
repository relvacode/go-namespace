package namespace

import (
	"fmt"
	"testing"
)

func TestStringNameSpace(t *testing.T) {
	_, err := StringNameSpace(nil, "")
	if err == nil {
		t.Fatal("no error given")
	}
}

type TestCase struct {
	Name      string
	Wants     string
	Namespace []string
	With      interface{}
}

type Embedded struct {
	Key string
}

type Child struct {
	Value string
}

type PassThrough struct {
	Container Child `ns:"-"`
}

type Rename struct {
	Container Child `ns:"rename"`
}

func TestNameSpace(t *testing.T) {
	cases := []TestCase{
		{
			Name:      "struct",
			Wants:     "Value",
			Namespace: []string{"Key"},
			With: struct {
				Key string
			}{Key: "Value"},
		},
		{
			Name:      "struct/struct",
			Wants:     "Value",
			Namespace: []string{"Key", "Child"},
			With: struct {
				Key struct {
					Child string
				}
			}{
				Key: struct {
					Child string
				}{
					Child: "Value",
				},
			},
		},
		{
			Name:      "struct/embedded",
			Wants:     "Value",
			Namespace: []string{"Key"},
			With: struct {
				Embedded
			}{
				Embedded: Embedded{Key: "Value"},
			},
		},
		{
			Name:      "map/string",
			Wants:     "Value",
			Namespace: []string{"Key"},
			With: map[string]string{
				"Key": "Value",
			},
		},
		{
			Name:      "map/interface",
			Wants:     "Value",
			Namespace: []string{"Key"},
			With: map[string]interface{}{
				"Key": "Value",
			},
		},
		{
			Name:      "map/interface/map/interface",
			Wants:     "Value",
			Namespace: []string{"Key", "Child"},
			With: map[string]interface{}{
				"Key": map[string]interface{}{
					"Child": "Value",
				},
			},
		},
		{
			Name:      "map/interface/struct",
			Wants:     "Value",
			Namespace: []string{"Key", "Child"},
			With: map[string]interface{}{
				"Key": struct {
					Child string
				}{
					Child: "Value",
				},
			},
		},
		{
			Name:      "struct/ns.-",
			Wants:     "Value",
			Namespace: []string{"Value"},
			With: PassThrough{
				Container: Child{
					Value: "Value",
				},
			},
		},
		{
			Name:      "struct/ns.rename",
			Wants:     "Value",
			Namespace: []string{"rename", "Value"},
			With: Rename{
				Container: Child{
					Value: "Value",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			v, err := NameSpace(c.With, c.Namespace...)
			if err != nil {
				t.Fatal(err)
			}
			if v.String() != c.Wants {
				t.Fatalf("wanted %s, got %s", c.Wants, v.String())
			}
		})
	}
}

func BenchmarkNameSpace_Struct(b *testing.B) {
	v := struct {
		Key struct {
			Key string
		}
	}{
		Key: struct {
			Key string
		}{
			Key: "Value",
		},
	}
	for i := 0; i < b.N; i++ {
		r, err := NameSpace(v, "Key", "Key")
		if err != nil {
			b.Fatal(err)
		}
		if r.String() != "Value" {
			b.Fatalf("wanted Value, got %s", r.String())
		}
	}
}

func BenchmarkNameSpace_Map(b *testing.B) {
	v := map[string]interface{}{
		"Key": map[string]interface{}{
			"Key": "Value",
		},
	}
	for i := 0; i < b.N; i++ {
		r, err := NameSpace(v, "Key", "Key")
		if err != nil {
			b.Fatal(err)
		}
		if r.String() != "Value" {
			b.Fatalf("wanted Value, got %s", r.String())
		}
	}
}

func Example() {
	// l is your value that you want to search its namespace for
	// It can be either a struct or map, or an interface to a struct or map.
	l := map[string]interface{}{
		"PrimaryKey": map[string]interface{}{
			"SecondaryKey": "MyValue",
		},
	}
	// Get a value by a full stop delimited string value.
	v, err := StringNameSpace(l, "PrimaryKey.SecondaryKey")
	// Or get a value by a slice of namespaces.
	v, err = NameSpace(l, "PrimaryKey", "SecondaryKey")

	if err != nil {
		// Do something with the error here
	}

	fmt.Println(v.String()) // Outputs: MyValue
}
