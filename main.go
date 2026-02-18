package main

import (
	"embed"

	"github.com/t4Linux/t4gfm/src/cmd"
)

var (
	//go:embed src/t4gfm_config/*
	content embed.FS
)

func main() {
	cmd.Run(content)
}
