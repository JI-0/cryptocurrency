package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/JI-0/private-cryptocurrency/blockchain"
	"github.com/JI-0/private-cryptocurrency/wallet"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	createChain -address ADDRESS <-- creates a blockchain")
	fmt.Println("	printChain <-- print the chain")
	fmt.Println("	createWallet <-- create a new wallet")
	fmt.Println("	listWallets <-- list addresses of all wallets")
	fmt.Println("	getBalance -address ADDRESS <-- get the balance for address")
	fmt.Println("	send -from FROM -to TO -amount AMOUNT <-- send amount from address to address")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) createChain(address string) {
	chain := blockchain.NewChain(address)
	chain.Database.Close()
	fmt.Println("Created chain")
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueChain("")
	defer chain.Database.Close()
	iterator := chain.Iterator()

	for {
		block := iterator.Next()
		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("POW: %s\n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.NewWallets()
	address := wallets.AddWallet()
	wallets.Save()
	println("New address: %s", address)
}

func (cli *CommandLine) listWallets() {
	wallets, _ := wallet.NewWallets()
	addresses := wallets.GetAllAddresses()
	for _, address := range addresses {
		println(address)
	}
}

func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueChain(address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueChain(from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Sent amount")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	createChainCmd := flag.NewFlagSet("createChain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	listWalletsCmd := flag.NewFlagSet("listWallets", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createChainAddress := createChainCmd.String("address", "", "The address")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createChain":
		if err := createChainCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "printChain":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "createWallet":
		if err := createWalletCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "listWallets":
		if err := listWalletsCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "getBalance":
		if err := getBalanceCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "send":
		if err := sendCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if createChainCmd.Parsed() {
		if *createChainAddress == "" {
			createChainCmd.Usage()
			runtime.Goexit()
		}
		cli.createChain(*createChainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listWalletsCmd.Parsed() {
		cli.listWallets()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}