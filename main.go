package main

import (
	"embed"
	"specfirst/cmd"
	"specfirst/internal/starter"
)

//go:embed starters/*
var startersFS embed.FS

var version = "0.5.0"

func main() {
	starter.EmbeddedFS = startersFS
	cmd.SetVersion(version)
	cmd.Execute()
}
