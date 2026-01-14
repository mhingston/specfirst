package main

import (
	"embed"
	"specfirst/cmd"
	"specfirst/internal/starter"
)

//go:embed starters/*
var startersFS embed.FS

var version = "v0.7.0"

func main() {
	starter.EmbeddedFS = startersFS
	cmd.SetVersion(version)
	cmd.Execute()
}
