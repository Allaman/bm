package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	createTables := `
		CREATE TABLE IF NOT EXISTS bookmarks (
				name TEXT PRIMARY KEY,
				url TEXT,
				archived INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS tags (
				name TEXT,
				tag TEXT,
				FOREIGN KEY(name) REFERENCES bookmarks(name) ON DELETE CASCADE,
				PRIMARY KEY(name, tag)
		);
	`
	_, err = db.Exec(createTables)
	if err != nil {
		return nil, err
	}

	// Migration: Add archived column if it doesn't exist
	_, err = db.Exec(`
		ALTER TABLE bookmarks ADD COLUMN archived INTEGER DEFAULT 0
	`)
	// Ignore error if column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

var ErrDuplicateURL = errors.New("name already exists")

func (r *SQLiteRepository) Add(b Bookmark) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && err == nil {
			err = fmt.Errorf("rollback error: %v", rbErr)
		}
	}()

	archived := 0
	if b.Archived {
		archived = 1
	}
	_, err = tx.Exec("INSERT INTO bookmarks (name, url, archived) VALUES (?, ?, ?)", b.Name, b.URL, archived)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateURL
		}
		return err
	}

	for _, tag := range b.Tags {
		_, err = tx.Exec("INSERT INTO tags (name, tag) VALUES (?, ?)", b.Name, strings.ToLower(tag))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) Del(name string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	result, err := tx.Exec("DELETE FROM bookmarks WHERE name = ?", name)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("rollback failed: %v", rbErr)
			return fmt.Errorf("delete failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("rollback failed: %v", rbErr)
			return fmt.Errorf("delete failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}
	if rows == 0 {
		return fmt.Errorf("bookmark not found")
	}

	_, err = tx.Exec("DELETE FROM tags WHERE name = ?", name)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("rollback failed: %v", rbErr)
			return fmt.Errorf("delete failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

func (r *SQLiteRepository) Update(b Bookmark, updateArchived bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && err == nil {
			err = fmt.Errorf("rollback error: %v", rbErr)
		}
	}()

	// Build dynamic update query
	updates := []string{}
	args := []any{}

	if b.URL != "" {
		updates = append(updates, "url = ?")
		args = append(args, b.URL)
	}

	if updateArchived {
		updates = append(updates, "archived = ?")
		archivedInt := 0
		if b.Archived {
			archivedInt = 1
		}
		args = append(args, archivedInt)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE bookmarks SET " + strings.Join(updates, ", ") + " WHERE name = ?"
	args = append(args, b.Name)

	result, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("bookmark not found")
	}

	// Only update tags if provided
	if len(b.Tags) > 0 {
		// Delete old tags
		_, err = tx.Exec("DELETE FROM tags WHERE name = ?", b.Name)
		if err != nil {
			return err
		}

		// Insert new tags
		for _, tag := range b.Tags {
			_, err = tx.Exec("INSERT INTO tags (name, tag) VALUES (?, ?)", b.Name, strings.ToLower(tag))
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) Ls(includeArchived bool) (Bookmarks, error) {
	query := `
        SELECT b.name, b.url, b.archived, GROUP_CONCAT(t.tag) as tags
        FROM bookmarks b
        LEFT JOIN tags t ON b.name = t.name`
	if !includeArchived {
		query += ` WHERE b.archived = 0`
	}
	query += ` GROUP BY b.name, b.url, b.archived`

	rows, err := r.db.Query(query)
	if err != nil {
		return Bookmarks{}, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var bookmarks Bookmarks
	for rows.Next() {
		var b Bookmark
		var tags sql.NullString
		var archived int
		if err := rows.Scan(&b.Name, &b.URL, &archived, &tags); err != nil {
			return Bookmarks{}, err
		}
		b.Archived = archived != 0
		if tags.Valid {
			b.Tags = strings.Split(tags.String, ",")
		}
		bookmarks.Bookmarks = append(bookmarks.Bookmarks, b)
	}
	return bookmarks, nil
}
