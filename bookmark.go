package main

type Repository interface {
	Add(bm Bookmark) error
	Del(title string) error
	Update(bm Bookmark, updateArchived bool, updateBrowser bool) error
	Ls(includeArchived bool) ([]Bookmark, error)
	Get(name string) (Bookmark, error)
	AddBrowser(b Browser) error
	DelBrowser(name string) error
	LsBrowsers() ([]Browser, error)
	GetBrowser(name string) (Browser, error)
}

type Bookmark struct {
	Name        string
	URL         string
	Tags        []string
	Archived    bool
	BrowserName string
}

type Browser struct {
	Name string
	Path string
	Args []string
}
