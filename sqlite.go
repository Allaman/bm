package main

import (
	"database/sql"
	"errors"
	"fmt"
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
				url TEXT
		);
		CREATE TABLE IF NOT EXISTS tags (
				name TEXT,
				tag TEXT,
				FOREIGN KEY(name) REFERENCES bookmarks(name) ON DELETE CASCADE,
				PRIMARY KEY(name, tag)
		);
	`
	_, err = db.Exec(createTables)
	return &SQLiteRepository{db: db}, err
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

	_, err = tx.Exec("INSERT INTO bookmarks (name, url) VALUES (?, ?)", b.Name, b.URL)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateURL
		}
		return err
	}

	for _, tag := range b.Tags {
		_, err = tx.Exec("INSERT INTO tags (name, tag) VALUES (?, ?)", b.Name, tag)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) Del(name string) error {
	result, err := r.db.Exec("DELETE FROM bookmarks WHERE name = ?", name)
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
	return nil
}

func (r *SQLiteRepository) Ls() (Bookmarks, error) {
	rows, err := r.db.Query(`
        SELECT b.name, b.url, GROUP_CONCAT(t.tag) as tags
        FROM bookmarks b
        LEFT JOIN tags t ON b.name = t.name
        GROUP BY b.name, b.url`)
	if err != nil {
		return Bookmarks{}, err
	}
	defer rows.Close()

	var bookmarks Bookmarks
	for rows.Next() {
		var b Bookmark
		var tags sql.NullString
		if err := rows.Scan(&b.Name, &b.URL, &tags); err != nil {
			return Bookmarks{}, err
		}
		if tags.Valid {
			b.Tags = strings.Split(tags.String, ",")
		}
		bookmarks.Bookmarks = append(bookmarks.Bookmarks, b)
	}
	return bookmarks, nil
}
