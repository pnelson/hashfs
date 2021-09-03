package hashfs

import (
	"embed"
	"io/fs"
	"testing"
)

//go:embed testdata
var testdata embed.FS

func TestHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want string
	}{
		{
			name: "testdata/base.ext",
			hash: "d476fb7be1b02ea9f66c797a0c11b11ff5db1bd702ff6c4720e455a301a501f1",
			want: "testdata/base.d476fb7be1b02ea9f66c797a0c11b11ff5db1bd702ff6c4720e455a301a501f1.ext",
		},
		{
			name: "testdata/noext",
			hash: "d9d4e730296e72377ae86529027f7defd5feecf4602a1f15f561cd4fae3644c5",
			want: "testdata/noext.d9d4e730296e72377ae86529027f7defd5feecf4602a1f15f561cd4fae3644c5",
		},
	}
	h := New(testdata)
	for _, tt := range tests {
		hash := h.Hash(tt.name)
		if hash != tt.hash {
			t.Errorf("Hash(%q)\nhave '%s'\nwant '%s'", tt.name, hash, tt.hash)
			continue
		}
		name := h.Name(tt.name)
		if name != tt.want {
			t.Errorf("Name(%q)\nhave '%s'\nwant '%s'", tt.name, name, tt.want)
		}
	}
}

func TestHashPathError(t *testing.T) {
	h := New(testdata)
	hash := h.Hash("not-found")
	if hash != "" {
		t.Errorf("should return an empty string")
	}
}

func TestOpen(t *testing.T) {
	h := New(testdata)
	tests := []string{
		"testdata/base.d476fb7be1b02ea9f66c797a0c11b11ff5db1bd702ff6c4720e455a301a501f1.ext",
		"testdata/noext.d9d4e730296e72377ae86529027f7defd5feecf4602a1f15f561cd4fae3644c5",
	}
	for _, tt := range tests {
		f, err := h.Open(tt)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		err = f.Close()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestOpenPathError(t *testing.T) {
	h := New(testdata)
	tests := []string{
		"testdata/base.ext",
		"testdata/base.8888888888888888888888888888888888888888888888888888888888888888.ext",
		"testdata/noext",
		"testdata/noext.8888888888888888888888888888888888888888888888888888888888888888",
	}
	for _, tt := range tests {
		_, err := h.Open(tt)
		if err == nil {
			t.Errorf("Open(%q) should error", tt)
			continue
		}
		_, ok := err.(*fs.PathError)
		if !ok {
			t.Errorf("Open(%q) should yield a *fs.PathError", tt)
		}
	}
}
