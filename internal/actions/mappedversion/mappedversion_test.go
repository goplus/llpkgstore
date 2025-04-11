package mappedversion

import "testing"

func TestMappedVersion(t *testing.T) {
	t.Run("invalid-1", func(t *testing.T) {
		_, _, err := From("cjson").Parse()
		if err == nil {
			t.Errorf("unpexted behavior: no error")
		}
	})
	t.Run("invalid-2", func(t *testing.T) {
		_, _, err := From("cjson/").Parse()
		if err == nil {
			t.Errorf("unpexted behavior: no error")
		}
	})

	t.Run("invalid-3", func(t *testing.T) {
		_, _, err := From("cjson/1.7.18").Parse()
		if err == nil {
			t.Errorf("unpexted behavior: no error")
		}
	})

	t.Run("valid", func(t *testing.T) {
		clib, version, err := From("cjson/v1.0.0").Parse()
		if err != nil {
			t.Errorf("unpexted error: %v", err)
		}
		if clib != "cjson" {
			t.Errorf("unpexted clib: want %s got %s", "cjson", clib)
		}
		if version != "v1.0.0" {
			t.Errorf("unpexted clib: want %s got %s", "v1.0.0", version)
		}
	})
}
