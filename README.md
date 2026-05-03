# BM

<h1 align="center">bm 🗺️</h1>

<div align="center">
  <p>
    <img src="https://github.com/Allaman/bm/actions/workflows/release.yaml/badge.svg" alt="Release"/>
    <img src="https://img.shields.io/github/repo-size/Allaman/bm" alt="size"/>
    <img src="https://img.shields.io/github/issues/Allaman/bm" alt="issues"/>
    <img src="https://img.shields.io/github/last-commit/Allaman/bm" alt="last commit"/>
    <img src="https://img.shields.io/github/license/Allaman/bm" alt="license"/>
    <img src="https://img.shields.io/github/v/release/Allaman/bm?sort=semver" alt="last release"/>
  </p>
 <em>Minimal bookmark management CLI.</em>
</div>

![screenshot](./screenshot.png)

Above command is in my [dotfiles](https://github.com/Allaman/dots/blob/main/dot_local/bin/executable_search-bookmark.sh)

## Why

- For fun
- Very minimalistic
- Platform and browser independent (I run various browser (profiles))

## Add a bookmark

Name must be unique as it is used as primary key.

```sh
bm [--path bookmarks.sqlite] add --url https://www.google.com --name Google [--tags foo bar --archive --browser zen-work]
```

## List bookmarks

```sh
bm [--path bookmarks.sqlite] ls [-s ";" -c -a -t -b]
```

## Delete bookmark

```sh
bm [--path bookmarks.sqlite] del --name Google
```

## Update bookmark

```sh
bm [--path bookmarks.sqlite] upd --url https://www.google2.com --name Google [--tags foo bar --unarchive --browser zen-work]
```

Pass `--browser ""` to remove the browser association from a bookmark.

## Open a bookmark

Opens the URL in the bookmark's configured browser profile. Falls back to the system default (`open`/`xdg-open`) when no profile is set.

```sh
bm [--path bookmarks.sqlite] open --name Google
```

## Browser profiles

Browser profiles map a name to a binary and its arguments. Each `-a` flag is one argument passed directly to the binary, so arguments with spaces are handled correctly.

```sh
# Add a profile
bm [--path bookmarks.sqlite] browser add --name zen-work --binary /Applications/Zen.app/Contents/MacOS/zen --args=-p --args=work

# List profiles
bm [--path bookmarks.sqlite] browser ls

# Delete a profile (bookmarks using it will have their association cleared)
bm [--path bookmarks.sqlite] browser del --name zen-work
```

## Search a bookmark

`bm` does not come with a search included. There are better tools out there that can handle this, e.g. [fzf](https://github.com/junegunn/fzf). Above, right under the screenshot I linked the script that calls bm.

This script is mapped to a shortcut in [skhd](https://github.com/koekeishiya/skhd):

```sh
cmd - b : open -n /Applications/Ghostty.app --args --title=bm --command="$HOME/.local/bin/search-bookmark.sh"
```

## Synopsis

```sh
bm --help
Usage: bm <command> [flags]

A minimal bookmarking management CLI

Flags:
  -h, --help                  Show context-sensitive help.
  -p, --path="./bm.sqlite"    Path to the sqlite database

Commands:
  add --url=STRING --name=STRING [flags]
    Add a new bookmark

  del --name=STRING [flags]
    Delete a bookmark

  ls [flags]
    List all bookmarks

  upd --name=STRING [flags]
    Update a bookmark

  open --name=STRING [flags]
    Open a bookmark in its configured browser

  browser add --name=STRING --path=STRING [flags]
    Add a browser profile

  browser del --name=STRING
    Delete a browser profile

  browser ls
    List browser profiles

  version [flags]
    Show version information

Run "bm <command> --help" for more information on a command.
```
