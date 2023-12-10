package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/JI-0/private-cryptocurrency/blockchain"
)

type CommandLine struct {
	chain *blockchain.Chain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	add -block BLOCK_DATA <-- add a block to the chain")
	fmt.Println("	print <-- print the chain")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) addBlock(data string) {
	cli.chain.AddBlock(data)
}

func (cli *CommandLine) printChain() {
	iterator := cli.chain.Iterator()

	for {
		block := iterator.Next()
		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		if err := addBlockCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "print":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	defer os.Exit(0)
	chain := blockchain.NewChain()
	defer chain.Database.Close()

	cli := CommandLine{chain}
	cli.run()
}
