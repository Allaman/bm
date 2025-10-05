package main

type Repository interface {
	Add(bm Bookmark) error
	Del(title string) error
	Update(bm Bookmark, updateArchived bool) error
	Ls(includeArchived bool) (Bookmarks, error)
}

type Bookmark struct {
	Name     string
	URL      string
	Tags     []string
	Archived bool
}

type Bookmarks struct {
	Bookmarks []Bookmark
}
