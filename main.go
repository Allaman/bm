package main

import (
	"log"

	"github.com/alecthomas/kong"
)

var Version = "dev"

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("bm"),
		kong.Description("A minimal bookmarking management CLI"),
		kong.UsageOnError())
	repository, err := NewSQLiteRepository(cli.Path)
	if err != nil {
		log.Fatal(err)
	}
	err = ctx.Run(&Context{Repository: repository})
	ctx.FatalIfErrorf(err)
	defer repository.db.Close()
}
