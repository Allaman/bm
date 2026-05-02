package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf8"
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
	Open    OpenCmd    `cmd:"" help:"Open a bookmark in its configured browser"`
	Browser BrowserCmd `cmd:"" help:"Manage browser profiles"`
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
	Browser  string   `short:"b" help:"Browser profile name to associate"`
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
	Browser   *string  `help:"Browser profile name; pass empty string to clear"`
}

type LsCmd struct {
	Separator       string `short:"s" default:"|" help:"Separator (one character)"`
	Colored         bool   `short:"c" default:"false" help:"Colored output"`
	IncludeArchived bool   `short:"a" default:"false" help:"Include archived bookmarks"`
	ShowTags        bool   `short:"t" default:"false" help:"Show tags"`
	ShowBrowser     bool   `short:"b" default:"false" help:"Show browser profile"`
}

type OpenCmd struct {
	Name string `short:"n" required:"" help:"Name of the bookmark to open"`
}

type BrowserCmd struct {
	Add BrowserAddCmd `cmd:"" help:"Add a browser profile"`
	Del BrowserDelCmd `cmd:"" help:"Delete a browser profile"`
	Ls  BrowserLsCmd  `cmd:"" help:"List browser profiles"`
}

type BrowserAddCmd struct {
	Name string `short:"n" required:"" help:"Profile name (e.g. zen-work)"`
	Path string `short:"x" required:"" help:"Absolute path to the browser binary"`
	Args []string `short:"a" help:"Arguments passed before the URL; repeat for each arg (e.g. -a -p -a work)"`
}

type BrowserDelCmd struct {
	Name string `short:"n" required:"" help:"Profile name to delete"`
}

type BrowserLsCmd struct{}

func (c *LsCmd) Validate() error {
	if utf8.RuneCountInString(c.Separator) != 1 {
		return fmt.Errorf("separator must be exactly one character, got %d", utf8.RuneCountInString(c.Separator))
	}
	return nil
}

func (c *UpdateCmd) Validate() error {
	if c.Archive && c.Unarchive {
		return fmt.Errorf("--archive and --unarchive are mutually exclusive")
	}
	if c.URL == "" && !c.Archive && !c.Unarchive && len(c.Tags) == 0 && c.Browser == nil {
		return fmt.Errorf("at least one of --url, --tags, --archive, --unarchive, or --browser must be specified")
	}
	return nil
}

func (c *AddCmd) Run(ctx *Context) error {
	return ctx.Repository.Add(Bookmark{
		Name:        c.Name,
		URL:         c.URL,
		Tags:        c.Tags,
		Archived:    c.Archived,
		BrowserName: c.Browser,
	})
}

func (c *LsCmd) Run(ctx *Context) error {
	bookmarks, err := ctx.Repository.Ls(c.IncludeArchived)
	if err != nil {
		return err
	}
	for _, bm := range bookmarks {
		tags := ""
		if c.ShowTags && len(bm.Tags) > 0 {
			tags = c.Separator + strings.Join(bm.Tags, ",")
		}
		browser := ""
		if c.ShowBrowser && bm.BrowserName != "" {
			browser = c.Separator + bm.BrowserName
		}

		if c.Colored {
			fmt.Printf("%s%s%s%s%s%s%s%s%s\n",
				ColorRed, bm.Name, ColorReset,
				c.Separator,
				ColorGreen, bm.URL, ColorReset,
				tags, browser)
		} else {
			fmt.Printf("%s%s%s%s%s\n", bm.Name, c.Separator, bm.URL, tags, browser)
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
	browserName := ""
	if c.Browser != nil {
		browserName = *c.Browser
	}
	return ctx.Repository.Update(
		Bookmark{Name: c.Name, URL: c.URL, Tags: c.Tags, Archived: archived, BrowserName: browserName},
		updateArchived,
		c.Browser != nil,
	)
}

func (c *OpenCmd) Run(ctx *Context) error {
	bm, err := ctx.Repository.Get(c.Name)
	if err != nil {
		return err
	}

	if bm.BrowserName == "" {
		return openDefault(bm.URL)
	}

	browser, err := ctx.Repository.GetBrowser(bm.BrowserName)
	if err != nil {
		return err
	}

	args := append(browser.Args, bm.URL)
	return exec.Command(browser.Path, args...).Start()
}

func (c *BrowserAddCmd) Run(ctx *Context) error {
	return ctx.Repository.AddBrowser(Browser{Name: c.Name, Path: c.Path, Args: c.Args})
}

func (c *BrowserDelCmd) Run(ctx *Context) error {
	return ctx.Repository.DelBrowser(c.Name)
}

func (c *BrowserLsCmd) Run(ctx *Context) error {
	browsers, err := ctx.Repository.LsBrowsers()
	if err != nil {
		return err
	}
	for _, b := range browsers {
		if len(b.Args) > 0 {
			fmt.Printf("%s\t%s %s\n", b.Name, b.Path, strings.Join(b.Args, " "))
		} else {
			fmt.Printf("%s\t%s\n", b.Name, b.Path)
		}
	}
	return nil
}

// openDefault opens a URL using the OS default browser.
func openDefault(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform %s: assign a browser profile to this bookmark", runtime.GOOS)
	}
	return cmd.Start()
}
