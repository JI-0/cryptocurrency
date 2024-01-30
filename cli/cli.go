package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/JI-0/private-cryptocurrency/blockchain"
	"github.com/JI-0/private-cryptocurrency/network"
	"github.com/JI-0/private-cryptocurrency/wallet"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	createChain -address ADDRESS <-- creates a blockchain")
	fmt.Println("	printChain <-- print the chain")
	fmt.Println("	reindexUTXOSet <-- reindexes the UTXO set of unspent transactions")
	fmt.Println("	createWallet <-- create a new wallet")
	fmt.Println("	listWallets <-- list addresses of all wallets")
	fmt.Println("	getBalance -address ADDRESS <-- get the balance for address")
	fmt.Println("	send -from FROM -to TO -amount AMOUNT -mine <-- send amount from address to address")
	fmt.Println("	startNode -miner ADDRESS <-- start a miner with address")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) createChain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		panic("address invalid")
	}
	chain := blockchain.NewChain(address, nodeID)
	chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Chain: chain}
	UTXOSet.Reindex()
	fmt.Println("Created chain")
}

func (cli *CommandLine) printChain(nodeID string) {
	chain := blockchain.ContinueChain(nodeID)
	defer chain.Database.Close()
	iterator := chain.Iterator()

	for {
		block := iterator.Next()
		fmt.Printf("Prev hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(chain, block, true)
		fmt.Printf("POW: %s\n", strconv.FormatBool(pow.Validate()))
		pow.Destroy()
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) reindexUTXOSet(nodeID string) {
	chain := blockchain.ContinueChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Chain: chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("There are %d transactions in the UTXO set.", count)
}

func (cli *CommandLine) createWallet(nodeID string) {
	wallets, _ := wallet.NewWallets()
	address := wallets.AddWallet()
	wallets.Save()
	println("New address: %s", address)
}

func (cli *CommandLine) listWallets(nodeID string) {
	wallets, _ := wallet.NewWallets()
	addresses := wallets.GetAllAddresses()
	for _, address := range addresses {
		println(address)
	}
}

func (cli *CommandLine) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		panic("address invalid")
	}
	chain := blockchain.ContinueChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Chain: chain}

	balance := 0
	publicKeyHash := wallet.Base58Decode([]byte(address))
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-wallet.ChecksumLen]
	UTXOs := UTXOSet.FindUTXO(publicKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mine bool) {
	if !wallet.ValidateAddress(from) || !wallet.ValidateAddress(to) {
		panic("address invalid")
	}
	chain := blockchain.ContinueChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockchain.UTXOSet{Chain: chain}

	wallets, err := wallet.NewWallets()
	if err != nil {
		panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewTransaction(&wallet, to, amount, &UTXOSet)
	if mine {
		cbTx := blockchain.CoinbaseTransaction(from, "")
		block := chain.MineBlock([]*blockchain.Transaction{tx, cbTx})
		UTXOSet.Update(block)
	} else {
		network.SendTransaction(network.KnownNodes[0], tx)
	}

	fmt.Println("Sent amount")
}

func (cli *CommandLine) StartNode(nodeId, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeId)
	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining. Address: ", minerAddress)
		} else {
			fmt.Println("Miner address error.")
		}
	}
	network.StartServer(nodeId, minerAddress)
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Println("NODE_ID is not set in the environment")
		runtime.Goexit()
	}

	createChainCmd := flag.NewFlagSet("createChain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	reindexUTXOSetCmd := flag.NewFlagSet("reindexUTXOSet", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	listWalletsCmd := flag.NewFlagSet("listWallets", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startNode", flag.ExitOnError)

	createChainAddress := createChainCmd.String("address", "", "The address")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable miner")

	switch os.Args[1] {
	case "createChain":
		if err := createChainCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "printChain":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			panic(err)
		}
	case "reindexUTXOSet":
		if err := reindexUTXOSetCmd.Parse(os.Args[2:]); err != nil {
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
	case "startNode":
		if err := startNodeCmd.Parse(os.Args[2:]); err != nil {
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
		cli.createChain(*createChainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if reindexUTXOSetCmd.Parsed() {
		cli.reindexUTXOSet(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listWalletsCmd.Parsed() {
		cli.listWallets(nodeID)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
}
