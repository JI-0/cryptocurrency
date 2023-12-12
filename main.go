package main

import (
	"os"

	"github.com/JI-0/private-cryptocurrency/cli"
)

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}
