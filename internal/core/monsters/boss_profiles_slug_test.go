package monsters

import (
	"testing"
)

func TestNameToSlug(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Ambushjaw Gator", "ambushjaw_gator"},
		{"The Cutahead Lizard King", "the_cutahead_lizard_king"},
		{"single", "single"},
		{"", ""},
		{"  spaces  ", "spaces"},
	}
	for _, tt := range tests {
		got := NameToSlug(tt.name)
		if got != tt.want {
			t.Errorf("NameToSlug(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
