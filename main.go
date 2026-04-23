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
	defer func() {
		if err := repository.db.Close(); err != nil {
			log.Printf("error closing database connection: %v", err)
		}
	}()
	err = ctx.Run(&Context{Repository: repository})
	ctx.FatalIfErrorf(err)
}
