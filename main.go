package main

import (
	"os"

	"github.com/JI-0/private-cryptocurrency/cli"
)

func main() {
	// pub, priv, _ := ed448.GenerateKey(nil)
	// os.WriteFile("keys/master_ed448_.pub", pub, 0644)
	// os.WriteFile("keys/master_ed448_.priv", priv, 0644)

	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}
