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
				Tags: []string{"Search", "web"},
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

	// Add a test bookmark
	bm := Bookmark{Name: "Test", URL: "https://test.com"}
	if err := repo.Add(bm); err != nil {
		t.Fatalf("failed to add test bookmark: %v", err)
	}

	// Add a tag for the test bookmark
	_, err := repo.db.Exec("INSERT INTO tags (name, tag) VALUES (?, ?)", "Test", "testtag")
	if err != nil {
		t.Fatalf("failed to add tag: %v", err)
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

			if !tt.wantError {
				// Verify bookmark was deleted
				var count int
				err = repo.db.QueryRow("SELECT COUNT(*) FROM bookmarks WHERE name = ?", tt.bookmark).Scan(&count)
				if err != nil {
					t.Fatalf("failed to query bookmarks: %v", err)
				}
				if count != 0 {
					t.Errorf("bookmark was not deleted from bookmarks table")
				}

				// Verify tag was deleted
				err = repo.db.QueryRow("SELECT COUNT(*) FROM tags WHERE name = ?", tt.bookmark).Scan(&count)
				if err != nil {
					t.Fatalf("failed to query tags table: %v", err)
				}
				if count != 0 {
					t.Errorf("tag entry was not deleted from tags table")
				}
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

func TestUpdate(t *testing.T) {
	repo := setupTestDB(t)

	initial := Bookmark{
		Name: "Google",
		URL:  "https://google.com",
		Tags: []string{"search"},
	}
	if err := repo.Add(initial); err != nil {
		t.Fatalf("failed to add initial bookmark: %v", err)
	}

	updated := Bookmark{
		Name: "Google",
		URL:  "https://google.co.uk",
		Tags: []string{"search", "uk"},
	}
	if err := repo.Update(updated); err != nil {
		t.Errorf("Update() error = %v", err)
	}

	result, err := repo.Ls()
	if err != nil {
		t.Fatalf("Ls() error = %v", err)
	}

	found := false
	for _, bm := range result.Bookmarks {
		if bm.Name == updated.Name {
			found = true
			if bm.URL != updated.URL {
				t.Errorf("got URL %q, want %q", bm.URL, updated.URL)
			}
			if !equalSlices(bm.Tags, updated.Tags) {
				t.Errorf("got tags %v, want %v", bm.Tags, updated.Tags)
			}
		}
	}
	if !found {
		t.Errorf("bookmark %q not found", updated.Name)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
