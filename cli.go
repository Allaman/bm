package main

import (
	"fmt"
	"strings"
)

const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
)

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
	URL      string   `short:"u" required:"" help:"URL of the bookmark"`
	Name     string   `short:"n" required:"" help:"Name of the bookmark (must be unique)"`
	Tags     []string `short:"t" help:"Tags for the bookmark"`
	Archived bool     `short:"a" default:"false" help:"Mark bookmark as archived"`
}

type DelCmd struct {
	Name string `short:"n" required:"" help:"Name to be deleted"`
}

type UpdateCmd struct {
	URL       string   `short:"u" help:"URL of the bookmark"`
	Name      string   `short:"n" required:"" help:"Name of the bookmark (must be unique)"`
	Tags      []string `short:"t" help:"Tags for the bookmark"`
	Archive   bool     `short:"a" help:"Mark bookmark as archived"`
	Unarchive bool     `help:"Mark bookmark as not archived"`
}

type LsCmd struct {
	Separator       string `short:"s" default:"|" help:"Separator (one character)"`
	Colored         bool   `short:"c" default:"false" help:"Colored output"`
	IncludeArchived bool   `short:"a" default:"false" help:"Include archived bookmarks"`
	ShowTags        bool   `short:"t" default:"false" help:"Show tags"`
}

func (c *LsCmd) Validate() error {
	if len(c.Separator) != 1 {
		return fmt.Errorf("argument must be exactly one character, got %d characters", len(c.Separator))
	}
	return nil
}

func (c *AddCmd) Run(ctx *Context) error {
	return ctx.Repository.Add(Bookmark{Name: c.Name, URL: c.URL, Tags: c.Tags, Archived: c.Archived})
}

func (c *LsCmd) Run(ctx *Context) error {
	bookmarks, err := ctx.Repository.Ls(c.IncludeArchived)
	if err != nil {
		return err
	}
	for _, bm := range bookmarks.Bookmarks {
		tags := ""
		if c.ShowTags && len(bm.Tags) > 0 {
			tags = c.Separator + strings.Join(bm.Tags, ",")
		}

		if c.Colored {
			fmt.Printf("%s%s%s%s%s%s%s%s\n", ColorRed, bm.Name, ColorReset, c.Separator, ColorGreen, bm.URL, ColorReset, tags)
		} else {
			fmt.Printf("%s%s%s%s\n", bm.Name, c.Separator, bm.URL, tags)
		}
	}
	return nil
}

func (c *DelCmd) Run(ctx *Context) error {
	return ctx.Repository.Del(c.Name)
}

func (c *UpdateCmd) Run(ctx *Context) error {
	updateArchived := c.Archive || c.Unarchive
	archived := c.Archive && !c.Unarchive
	return ctx.Repository.Update(Bookmark{Name: c.Name, URL: c.URL, Tags: c.Tags, Archived: archived}, updateArchived)
}
