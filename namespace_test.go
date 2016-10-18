package namespace

import "testing"

func TestGetString(t *testing.T) {
	_, err := GetString(nil, "")
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

func TestGet(t *testing.T) {
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
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			v, err := Get(c.With, c.Namespace...)
			if err != nil {
				t.Fatal(err)
			}
			if v.String() != c.Wants {
				t.Fatalf("wanted %s, got %s", c.Wants, v.String())
			}
		})
	}
}

func BenchmarkGet_Struct(b *testing.B) {
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
		r, err := Get(v, "Key", "Key")
		if err != nil {
			b.Fatal(err)
		}
		if r.String() != "Value" {
			b.Fatalf("wanted Value, got %s", r.String())
		}
	}
}

func BenchmarkGet_Map(b *testing.B) {
	v := map[string]interface{}{
		"Key": map[string]interface{}{
			"Key": "Value",
		},
	}
	for i := 0; i < b.N; i++ {
		r, err := Get(v, "Key", "Key")
		if err != nil {
			b.Fatal(err)
		}
		if r.String() != "Value" {
			b.Fatalf("wanted Value, got %s", r.String())
		}
	}
}
