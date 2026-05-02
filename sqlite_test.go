package main

import (
	"errors"
	"os"
	"slices"
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
			wantErr: ErrDuplicateName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Add(tt.bm)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Add() error = %v, want %v", err, tt.wantErr)
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

	result, err := repo.Ls(true)
	if err != nil {
		t.Fatalf("Ls() error = %v", err)
	}

	if len(result) != len(bookmarks) {
		t.Errorf("got %d bookmarks, want %d", len(result), len(bookmarks))
	}

	for _, want := range bookmarks {
		found := false
		for _, got := range result {
			if got.Name == want.Name {
				found = true
				if got.URL != want.URL {
					t.Errorf("got URL %q, want %q", got.URL, want.URL)
				}
				if !slices.Equal(got.Tags, want.Tags) {
					t.Errorf("got tags %v, want %v", got.Tags, want.Tags)
				}
			}
		}
		if !found {
			t.Errorf("bookmark %q not found", want.Name)
		}
	}
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
	if err := repo.Update(updated, false, false); err != nil {
		t.Errorf("Update() error = %v", err)
	}

	result, err := repo.Ls(true)
	if err != nil {
		t.Fatalf("Ls() error = %v", err)
	}

	found := false
	for _, bm := range result {
		if bm.Name == updated.Name {
			found = true
			if bm.URL != updated.URL {
				t.Errorf("got URL %q, want %q", bm.URL, updated.URL)
			}
			if !slices.Equal(bm.Tags, updated.Tags) {
				t.Errorf("got tags %v, want %v", bm.Tags, updated.Tags)
			}
		}
	}
	if !found {
		t.Errorf("bookmark %q not found", updated.Name)
	}
}

func TestArchivedFiltering(t *testing.T) {
	repo := setupTestDB(t)

	bookmarks := []Bookmark{
		{Name: "Active1", URL: "https://active1.com", Archived: false},
		{Name: "Active2", URL: "https://active2.com", Archived: false},
		{Name: "Archived1", URL: "https://archived1.com", Archived: true},
	}

	for _, bm := range bookmarks {
		if err := repo.Add(bm); err != nil {
			t.Fatalf("failed to add test bookmark: %v", err)
		}
	}

	t.Run("exclude archived", func(t *testing.T) {
		result, err := repo.Ls(false)
		if err != nil {
			t.Fatalf("Ls(false) error = %v", err)
		}

		if len(result) != 2 {
			t.Errorf("got %d bookmarks, want 2", len(result))
		}

		for _, bm := range result {
			if bm.Archived {
				t.Errorf("found archived bookmark %q in non-archived list", bm.Name)
			}
		}
	})

	t.Run("include archived", func(t *testing.T) {
		result, err := repo.Ls(true)
		if err != nil {
			t.Fatalf("Ls(true) error = %v", err)
		}

		if len(result) != 3 {
			t.Errorf("got %d bookmarks, want 3", len(result))
		}

		archivedCount := 0
		for _, bm := range result {
			if bm.Archived {
				archivedCount++
			}
		}
		if archivedCount != 1 {
			t.Errorf("got %d archived bookmarks, want 1", archivedCount)
		}
	})
}

func TestBrowser(t *testing.T) {
	repo := setupTestDB(t)

	zen := Browser{Name: "zen-work", Path: "/Applications/Zen.app/Contents/MacOS/zen", Args: []string{"-p", "work"}}

	t.Run("add browser", func(t *testing.T) {
		if err := repo.AddBrowser(zen); err != nil {
			t.Fatalf("AddBrowser() error = %v", err)
		}
	})

	t.Run("duplicate browser", func(t *testing.T) {
		if !errors.Is(repo.AddBrowser(zen), ErrDuplicateBrowser) {
			t.Error("expected ErrDuplicateBrowser")
		}
	})

	t.Run("list browsers", func(t *testing.T) {
		browsers, err := repo.LsBrowsers()
		if err != nil {
			t.Fatalf("LsBrowsers() error = %v", err)
		}
		if len(browsers) != 1 {
			t.Fatalf("got %d browsers, want 1", len(browsers))
		}
		if browsers[0].Name != zen.Name || browsers[0].Path != zen.Path || !slices.Equal(browsers[0].Args, zen.Args) {
			t.Errorf("got %+v, want %+v", browsers[0], zen)
		}
	})

	t.Run("get browser", func(t *testing.T) {
		b, err := repo.GetBrowser(zen.Name)
		if err != nil {
			t.Fatalf("GetBrowser() error = %v", err)
		}
		if b.Path != zen.Path || !slices.Equal(b.Args, zen.Args) {
			t.Errorf("got %+v, want %+v", b, zen)
		}
	})

	t.Run("get nonexistent browser", func(t *testing.T) {
		if _, err := repo.GetBrowser("nope"); err == nil {
			t.Error("expected error for nonexistent browser")
		}
	})

	t.Run("bookmark with browser", func(t *testing.T) {
		bm := Bookmark{Name: "Work Google", URL: "https://google.com", BrowserName: zen.Name}
		if err := repo.Add(bm); err != nil {
			t.Fatalf("Add() error = %v", err)
		}

		got, err := repo.Get(bm.Name)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.BrowserName != zen.Name {
			t.Errorf("got BrowserName %q, want %q", got.BrowserName, zen.Name)
		}
	})

	t.Run("bookmark with invalid browser", func(t *testing.T) {
		bm := Bookmark{Name: "Bad BM", URL: "https://example.com", BrowserName: "nonexistent"}
		if err := repo.Add(bm); err == nil {
			t.Error("expected error for nonexistent browser profile")
		}
	})

	t.Run("update bookmark browser", func(t *testing.T) {
		bm := Bookmark{Name: "Work Google", BrowserName: ""}
		if err := repo.Update(bm, false, true); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		got, err := repo.Get("Work Google")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.BrowserName != "" {
			t.Errorf("expected BrowserName to be cleared, got %q", got.BrowserName)
		}
	})

	t.Run("delete browser cascades to bookmarks", func(t *testing.T) {
		// Re-associate the bookmark with the browser
		bm := Bookmark{Name: "Work Google", BrowserName: zen.Name}
		if err := repo.Update(bm, false, true); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		// Delete the browser; ON DELETE SET NULL should clear the FK
		if err := repo.DelBrowser(zen.Name); err != nil {
			t.Fatalf("DelBrowser() error = %v", err)
		}

		got, err := repo.Get("Work Google")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.BrowserName != "" {
			t.Errorf("expected BrowserName to be NULL after browser deletion, got %q", got.BrowserName)
		}
	})

	t.Run("delete nonexistent browser", func(t *testing.T) {
		if err := repo.DelBrowser("nope"); err == nil {
			t.Error("expected error for nonexistent browser")
		}
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
