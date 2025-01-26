package main

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) *SQLiteRepository {
	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return repo
}

func TestAdd(t *testing.T) {
	repo := setupTestDB(t)

	tests := []struct {
		name    string
		bm      Bookmark
		wantErr error
	}{
		{
			name: "valid bookmark",
			bm: Bookmark{
				Name: "Google",
				URL:  "https://google.com",
				Tags: []string{"search", "web"},
			},
			wantErr: nil,
		},
		{
			name: "duplicate name",
			bm: Bookmark{
				Name: "Google",
				URL:  "https://different.com",
			},
			wantErr: ErrDuplicateURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Add(tt.bm)
			if err != tt.wantErr {
				if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
					t.Errorf("Add() error = %v, want %v", err, tt.wantErr)
				} else if (err == nil) != (tt.wantErr == nil) {
					t.Errorf("Add() error = %v, want %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDel(t *testing.T) {
	repo := setupTestDB(t)

	bm := Bookmark{Name: "Test", URL: "https://test.com"}
	if err := repo.Add(bm); err != nil {
		t.Fatalf("failed to add test bookmark: %v", err)
	}

	tests := []struct {
		name      string
		bookmark  string
		wantError bool
	}{
		{
			name:      "existing bookmark",
			bookmark:  "Test",
			wantError: false,
		},
		{
			name:      "non-existing bookmark",
			bookmark:  "NonExistent",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Del(tt.bookmark)
			if (err != nil) != tt.wantError {
				t.Errorf("Del() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLs(t *testing.T) {
	repo := setupTestDB(t)

	bookmarks := []Bookmark{
		{Name: "Google", URL: "https://google.com", Tags: []string{"search"}},
		{Name: "GitHub", URL: "https://github.com", Tags: []string{"dev", "git"}},
	}

	for _, bm := range bookmarks {
		if err := repo.Add(bm); err != nil {
			t.Fatalf("failed to add test bookmark: %v", err)
		}
	}

	result, err := repo.Ls()
	if err != nil {
		t.Fatalf("Ls() error = %v", err)
	}

	if len(result.Bookmarks) != len(bookmarks) {
		t.Errorf("got %d bookmarks, want %d", len(result.Bookmarks), len(bookmarks))
	}

	for _, want := range bookmarks {
		found := false
		for _, got := range result.Bookmarks {
			if got.Name == want.Name {
				found = true
				if got.URL != want.URL {
					t.Errorf("got URL %q, want %q", got.URL, want.URL)
				}
				if !equalSlices(got.Tags, want.Tags) {
					t.Errorf("got tags %v, want %v", got.Tags, want.Tags)
				}
			}
		}
		if !found {
			t.Errorf("bookmark %q not found", want.Name)
		}
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
