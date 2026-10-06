## What's Changed in 0.15.1
* chore: Update .gitignore by @Allaman
* fix: allow updating only a bookmark's tags by @Allaman
* fix: report rollback errors via named error returns by @Allaman
* fix: escape URI characters in the database path by @Allaman
* fix: skip opening the database for the version command by @Allaman
* fix: ignore duplicate tags that differ only in case by @Allaman
* chore(deps): Bump github/codeql-action from 4.38.0 to 4.38.2 by @Allaman in [#18](https://github.com/Allaman/bm/pull/18)
* chore(deps): Bump github/codeql-action from 4.37.9 to 4.38.0 by @Allaman in [#16](https://github.com/Allaman/bm/pull/16)
* chore(deps): Bump github/codeql-action from 4.37.8 to 4.37.9 by @Allaman in [#15](https://github.com/Allaman/bm/pull/15)
* chore(deps): Bump github/codeql-action from 4.37.7 to 4.37.8 by @Allaman in [#14](https://github.com/Allaman/bm/pull/14)
* chore(deps): Bump github/codeql-action from 4.37.6 to 4.37.7 by @Allaman in [#13](https://github.com/Allaman/bm/pull/13)
* chore(deps): Bump github/codeql-action from 4.37.4 to 4.37.6 by @Allaman in [#12](https://github.com/Allaman/bm/pull/12)
* chore(deps): Bump github/codeql-action from 4 to 4.37.4 by @Allaman in [#11](https://github.com/Allaman/bm/pull/11)
* chore(deps): Bump actions/setup-go from 6 to 7 by @Allaman in [#9](https://github.com/Allaman/bm/pull/9)
* chore(deps): Bump actions/checkout from 6 to 7 by @Allaman in [#8](https://github.com/Allaman/bm/pull/8)
* chore: markdownlint shenanigens by @Allaman
* chore: Update markdownlint config by @Allaman
* docs: Update CHANGELOG by @Allaman

**Full Changelog**: https://github.com/Allaman/bm/compare/0.15.0...0.15.1

## What's Changed in 0.15.0
* fix: Add wait and debug log to open command by @Allaman
* docs: Update CHANGELOG by @Allaman

**Full Changelog**: https://github.com/Allaman/bm/compare/0.14.0...0.15.0

## What's Changed in 0.14.0
* fix: bookmark arguments by @Allaman
* chore: Update gitignore by @Allaman
* feat: add browser profile support and open command by @Allaman
* fix: enforce SQLite FK constraints via DSN and single connection by @Allaman
* chore: Add git-cliff config by @Allaman
* ci: Add changelog by @Allaman
* ci: Add govulncheck by @Allaman
* test: use errors.Is and slices.Equal, remove equalSlices by @Allaman
* feat: add UpdateCmd.Validate for flag validation by @Allaman
* fix: use utf8.RuneCountInString for separator validation by @Allaman
* fix: check rows.Err() after iteration in Ls by @Allaman
* fix: use consistent deferred rollback pattern across all transactions by @Allaman
* fix: rename ErrDuplicateURL to ErrDuplicateName by @Allaman
* fix: enable foreign_keys PRAGMA and improve migration detection by @Allaman
* refactor: remove Bookmarks wrapper type, use []Bookmark directly by @Allaman
* fix: move defer db.Close before ctx.Run by @Allaman
* chore(deps): Bump goreleaser/goreleaser-action from 6 to 7 by @Allaman in [#7](https://github.com/Allaman/bm/pull/7)
* chore(deps): Bump actions/checkout from 5 to 6 by @Allaman in [#6](https://github.com/Allaman/bm/pull/6)
* chore(deps): Bump golangci/golangci-lint-action from 8 to 9 by @Allaman in [#5](https://github.com/Allaman/bm/pull/5)

**Full Changelog**: https://github.com/Allaman/bm/compare/0.13.0...0.14.0

## What's Changed in 0.13.0
* docs: Update README by @Allaman
* feat: Add show tags flag by @Allaman
* chore: clean up go.mod by @Allaman
* feat: Add (un)archive feature by @Allaman
* ci: Deprecated archives.format in goreleaser by @Allaman
* chore(deps): Bump actions/setup-go from 5 to 6 by @Allaman in [#4](https://github.com/Allaman/bm/pull/4)
* chore(deps): Bump actions/checkout from 4 to 5 by @Allaman in [#3](https://github.com/Allaman/bm/pull/3)
* chore(deps): Bump golangci/golangci-lint-action from 7 to 8 by @Allaman in [#2](https://github.com/Allaman/bm/pull/2)
* chore(deps): Bump golangci/golangci-lint-action from 6 to 7 by @Allaman in [#1](https://github.com/Allaman/bm/pull/1)

## New Contributors
* @dependabot[bot] made their first contribution

**Full Changelog**: https://github.com/Allaman/bm/compare/0.12.0...0.13.0

## What's Changed in 0.12.0
* chore: Fix lint errors by @Allaman
* fix: Also delete tag when deleting a bookmark by @Allaman
* docs: Update README by @Allaman
* docs: Add screenshot by @Allaman

**Full Changelog**: https://github.com/Allaman/bm/compare/0.11.0...0.12.0

## What's Changed in 0.11.0
* chore: Update Taskfile by @Allaman
* feat(sqlite)!: Always to lower tags by @Allaman
* feat(ls): Add option for basic colored output by @Allaman
* feat(ls): Add option for setting separator character by @Allaman

**Full Changelog**: https://github.com/Allaman/bm/compare/0.10.0...0.11.0

## What's Changed in 0.10.0
* feat: Add update bookmark feature by @Allaman

**Full Changelog**: https://github.com/Allaman/bm/compare/0.9.0...0.10.0

## What's Changed in 0.9.0
* docs: Fix typo by @Allaman
* docs: Update README by @Allaman
* ci: Fix Golang version by @Allaman
* initial commit by @Allaman

## New Contributors
* @Allaman made their first contribution


