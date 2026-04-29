package main

import (
	"reflect"
	"sort"
	"testing"
)

func sortStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func TestProcessTags(t *testing.T) {
	app := &application{}

	tests := []struct {
		name         string
		selectedTags []string
		customTags   string
		want         []string
	}{
		{
			name:         "deduplicates selected and custom",
			selectedTags: []string{"go", "web", "go"},
			customTags:   "web, sqlite, go",
			want:         []string{"go", "web", "sqlite"},
		},
		{
			name:         "ignores empty selected values",
			selectedTags: []string{"", " ", "api"},
			customTags:   "",
			want:         []string{"api"},
		},
		{
			name:         "trims custom tags",
			selectedTags: []string{"forum"},
			customTags:   " auth  , posts, comments ",
			want:         []string{"forum", "auth", "posts", "comments"},
		},
		{
			name:         "empty input returns empty output",
			selectedTags: nil,
			customTags:   "",
			want:         []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := app.processTags(tc.selectedTags, tc.customTags)
			if !reflect.DeepEqual(sortStrings(got), sortStrings(tc.want)) {
				t.Fatalf("processTags() = %v, want %v", got, tc.want)
			}
		})
	}
}
