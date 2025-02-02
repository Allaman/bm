package main

import "fmt"

type CLI struct {
	Add     AddCmd     `cmd:"" help:"Add a new bookmark"`
	Del     DelCmd     `cmd:"" help:"Delete a bookmark"`
	Ls      LsCmd      `cmd:"" help:"List all bookmarks"`
	Upd     UpdateCmd  `cmd:"" help:"Update a bookmark"`
	Path    string     `short:"p" default:"./bm.sqlite" help:"Path to the sqlite database"`
	Version VersionCmd `cmd:"" help:"Show version information"`
}

type VersionCmd struct {
	Version string
}

func (c *VersionCmd) Run() error {
	fmt.Println(Version)
	return nil
}

type Context struct {
	Repository Repository
}

type AddCmd struct {
	URL  string   `short:"u" required:"" help:"URL of the bookmark"`
	Name string   `short:"n" required:"" help:"Name of the bookmark (must be unique)"`
	Tags []string `short:"t" help:"Tags for the bookmark"`
}

type DelCmd struct {
	Name string `short:"n" required:"" help:"Name to be deleted"`
}

type UpdateCmd struct {
	URL  string   `short:"u" required:"" help:"URL of the bookmark"`
	Name string   `short:"n" required:"" help:"Name of the bookmark (must be unique)"`
	Tags []string `short:"t" help:"Tags for the bookmark"`
}

type LsCmd struct {
	Separator string `short:"s" default:"|" help:"Separator (one character)"`
}

func (c *LsCmd) Validate() error {
	if len(c.Separator) != 1 {
		return fmt.Errorf("argument must be exactly one character, got %d characters", len(c.Separator))
	}
	return nil
}

func (c *AddCmd) Run(ctx *Context) error {
	return ctx.Repository.Add(Bookmark{Name: c.Name, URL: c.URL, Tags: c.Tags})
}

func (c *LsCmd) Run(ctx *Context) error {
	bookmarks, err := ctx.Repository.Ls()
	if err != nil {
		return err
	}
	for _, bm := range bookmarks.Bookmarks {
		fmt.Printf("%s%s%s\n", bm.Name, c.Separator, bm.URL)
	}
	return nil
}

func (c *DelCmd) Run(ctx *Context) error {
	return ctx.Repository.Del(c.Name)
}

func (c *UpdateCmd) Run(ctx *Context) error {
	return ctx.Repository.Update(Bookmark{Name: c.Name, URL: c.URL, Tags: c.Tags})
}
