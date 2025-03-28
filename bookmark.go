package main

type Repository interface {
	Add(bm Bookmark) error
	Del(title string) error
	Update(bm Bookmark) error
	Ls() (Bookmarks, error)
}

type Bookmark struct {
	Name string
	URL  string
	Tags []string
}

type Bookmarks struct {
	Bookmarks []Bookmark
}
