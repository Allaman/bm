package main

import (
	"database/sql"
	"encoding/json"
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
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	createTables := `
		CREATE TABLE IF NOT EXISTS browsers (
				name TEXT PRIMARY KEY,
				path TEXT NOT NULL,
				args TEXT DEFAULT ''
		);
		CREATE TABLE IF NOT EXISTS bookmarks (
				name TEXT PRIMARY KEY,
				url TEXT,
				archived INTEGER DEFAULT 0,
				browser TEXT REFERENCES browsers(name) ON DELETE SET NULL
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

	// Migration: Add archived column if it doesn't exist (pre-archive DBs).
	if err = addColumnIfMissing(db, "bookmarks", "archived", "INTEGER DEFAULT 0"); err != nil {
		return nil, err
	}

	// Migration: Add browser column if it doesn't exist.
	if err = addColumnIfMissing(db, "bookmarks", "browser", "TEXT REFERENCES browsers(name) ON DELETE SET NULL"); err != nil {
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

func addColumnIfMissing(db *sql.DB, table, column, definition string) error {
	var count int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s'`, table, column)
	if err := db.QueryRow(q).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		_, err := db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, definition))
		return err
	}
	return nil
}

var (
	ErrDuplicateName    = errors.New("name already exists")
	ErrDuplicateBrowser = errors.New("browser profile already exists")
)

func (r *SQLiteRepository) Add(b Bookmark) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) && err == nil {
			err = fmt.Errorf("rollback error: %v", rbErr)
		}
	}()

	archived := 0
	if b.Archived {
		archived = 1
	}
	var browserArg any
	if b.BrowserName != "" {
		browserArg = b.BrowserName
	}
	_, err = tx.Exec(
		"INSERT INTO bookmarks (name, url, archived, browser) VALUES (?, ?, ?, ?)",
		b.Name, b.URL, archived, browserArg,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateName
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return fmt.Errorf("browser profile %q not found", b.BrowserName)
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
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) && err == nil {
			err = fmt.Errorf("rollback error: %v", rbErr)
		}
	}()

	var result sql.Result
	result, err = tx.Exec("DELETE FROM bookmarks WHERE name = ?", name)
	if err != nil {
		return err
	}

	var rows int64
	rows, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		err = fmt.Errorf("bookmark not found")
		return err
	}

	return tx.Commit()
}

func (r *SQLiteRepository) Update(b Bookmark, updateArchived bool, updateBrowser bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) && err == nil {
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

	if updateBrowser {
		updates = append(updates, "browser = ?")
		if b.BrowserName != "" {
			args = append(args, b.BrowserName)
		} else {
			args = append(args, nil) // NULL clears the association
		}
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE bookmarks SET " + strings.Join(updates, ", ") + " WHERE name = ?"
	args = append(args, b.Name)

	result, err := tx.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return fmt.Errorf("browser profile %q not found", b.BrowserName)
		}
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
		_, err = tx.Exec("DELETE FROM tags WHERE name = ?", b.Name)
		if err != nil {
			return err
		}

		for _, tag := range b.Tags {
			_, err = tx.Exec("INSERT INTO tags (name, tag) VALUES (?, ?)", b.Name, strings.ToLower(tag))
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *SQLiteRepository) Ls(includeArchived bool) ([]Bookmark, error) {
	query := `
        SELECT b.name, b.url, b.archived, GROUP_CONCAT(t.tag) as tags, b.browser
        FROM bookmarks b
        LEFT JOIN tags t ON b.name = t.name`
	if !includeArchived {
		query += ` WHERE b.archived = 0`
	}
	query += ` GROUP BY b.name, b.url, b.archived, b.browser`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		var tags sql.NullString
		var archived int
		var browserName sql.NullString
		if err := rows.Scan(&b.Name, &b.URL, &archived, &tags, &browserName); err != nil {
			return nil, err
		}
		b.Archived = archived != 0
		if tags.Valid {
			b.Tags = strings.Split(tags.String, ",")
		}
		if browserName.Valid {
			b.BrowserName = browserName.String
		}
		bookmarks = append(bookmarks, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookmarks, nil
}

func (r *SQLiteRepository) Get(name string) (Bookmark, error) {
	var b Bookmark
	var tags sql.NullString
	var archived int
	var browserName sql.NullString
	err := r.db.QueryRow(`
		SELECT b.name, b.url, b.archived, GROUP_CONCAT(t.tag) as tags, b.browser
		FROM bookmarks b
		LEFT JOIN tags t ON b.name = t.name
		WHERE b.name = ?
		GROUP BY b.name`, name).Scan(&b.Name, &b.URL, &archived, &tags, &browserName)
	if errors.Is(err, sql.ErrNoRows) {
		return Bookmark{}, fmt.Errorf("bookmark %q not found", name)
	}
	if err != nil {
		return Bookmark{}, err
	}
	b.Archived = archived != 0
	if tags.Valid {
		b.Tags = strings.Split(tags.String, ",")
	}
	if browserName.Valid {
		b.BrowserName = browserName.String
	}
	return b, nil
}

func (r *SQLiteRepository) AddBrowser(b Browser) error {
	argsJSON, err := json.Marshal(b.Args)
	if err != nil {
		return fmt.Errorf("encoding args: %w", err)
	}
	_, err = r.db.Exec(
		"INSERT INTO browsers (name, path, args) VALUES (?, ?, ?)",
		b.Name, b.Path, string(argsJSON),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrDuplicateBrowser
	}
	return err
}

func (r *SQLiteRepository) DelBrowser(name string) error {
	result, err := r.db.Exec("DELETE FROM browsers WHERE name = ?", name)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("browser profile %q not found", name)
	}
	return nil
}

func (r *SQLiteRepository) LsBrowsers() ([]Browser, error) {
	rows, err := r.db.Query("SELECT name, path, args FROM browsers ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var browsers []Browser
	for rows.Next() {
		var b Browser
		var argsJSON string
		if err := rows.Scan(&b.Name, &b.Path, &argsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(argsJSON), &b.Args); err != nil {
			return nil, fmt.Errorf("decoding args for %q: %w", b.Name, err)
		}
		browsers = append(browsers, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return browsers, nil
}

func (r *SQLiteRepository) GetBrowser(name string) (Browser, error) {
	var b Browser
	var argsJSON string
	err := r.db.QueryRow(
		"SELECT name, path, args FROM browsers WHERE name = ?", name,
	).Scan(&b.Name, &b.Path, &argsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Browser{}, fmt.Errorf("browser profile %q not found", name)
	}
	if err != nil {
		return Browser{}, err
	}
	if err := json.Unmarshal([]byte(argsJSON), &b.Args); err != nil {
		return Browser{}, fmt.Errorf("decoding args for %q: %w", name, err)
	}
	return b, nil
}
