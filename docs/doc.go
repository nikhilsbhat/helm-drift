package main

import (
	"log"

	"github.com/nikhilsbhat/helm-drift/cmd"
	"github.com/spf13/cobra/doc"
)

//go:generate go run github.com/nikhilsbhat/helm-drift/docs
func main() {
	commands := cmd.SetDriftCommands()

	if err := doc.GenMarkdownTree(commands, "doc"); err != nil {
		log.Fatal(err)
	}
}
