package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"

	"github.com/JI-0/private-cryptocurrency/blockchain"
	"github.com/vrecan/death/v3"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 6
)

var (
	nodeAddress     string
	minerAddress    string
	KnownNodes      = []string{"localhost:3000"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]blockchain.Transaction)
)

type Address struct {
	AddressList []string
}

type Block struct {
	AddressFrom string
	Block       []byte
}

type GetBlocks struct {
	AddressFrom string
}

type GetData struct {
	AddressFrom string
	Type        string
	ID          []byte
}

type Inventory struct {
	AddressFrom string
	Type        string
	Items       [][]byte
}

type Transaction struct {
	AddressFrom string
	Transaction []byte
}

type Version struct {
	AddressFrom string
	Version     int
	TopHeight   int
}

func SendAddress(address string) {
	nodes := Address{KnownNodes}
	nodes.AddressList = append(nodes.AddressList, nodeAddress)
	payload := GobEncode(nodes)
	requests := append(CmdToBytes("adr"), payload...)
	SendData(address, requests)
}

func SendBlock(address string, b *blockchain.Block) {
	data := Block{nodeAddress, b.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("blk"), payload...)
	SendData(address, request)
}

func SendInventory(address, kind string, items [][]byte) {
	inventory := Inventory{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CmdToBytes("inv"), payload...)
	SendData(address, request)
}

func SendTransaction(address string, tnx *blockchain.Transaction) {
	data := Transaction{nodeAddress, tnx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tnx"), payload...)
	SendData(address, request)
}

func SendVersion(address string, chain *blockchain.Chain) {
	topHeight := chain.GetTopHeight()
	payload := GobEncode(Version{nodeAddress, topHeight, version})
	request := append(CmdToBytes("vsn"), payload...)
	SendData(address, request)
}

func SendGetBlocks(address string) {
	payload := GobEncode(GetBlocks{nodeAddress})
	request := append(CmdToBytes("gbk"), payload...)
	SendData(address, request)
}

func SendGetData(address, kind string, id []byte) {
	payload := GobEncode(GetData{nodeAddress, kind, id})
	request := append(CmdToBytes("gdt"), payload...)
	SendData(address, request)
}

func SendData(address string, data []byte) {
	conn, err := net.Dial(protocol, address)

	if err != nil {
		fmt.Printf("%s is not available\n", address)
		var updatedNodes []string
		for _, node := range KnownNodes {
			if node != address {
				updatedNodes = append(updatedNodes, node)
			}
		}
		KnownNodes = updatedNodes
		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
	}
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(data); err != nil {
		panic(err)
	}

	return buff.Bytes()
}

func HandleAddress(request []byte) {
	var buffer bytes.Buffer
	var payload Address

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}
	KnownNodes = append(KnownNodes, payload.AddressList...)
	fmt.Printf("%d known nodes\n", len(KnownNodes))
	RequestBlocks()
}

func HandleBlock(request []byte, c *blockchain.Chain) {
	var buffer bytes.Buffer
	var payload Block
	var block *blockchain.Block

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}
	blockData := payload.Block
	block = block.Deserialize(blockData)
	fmt.Println("New block received")
	c.AddBlock(block)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddressFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{Chain: c}
		UTXOSet.Reindex()
	}
}

func HandleGetBlocks(request []byte, c *blockchain.Chain) {
	var buffer bytes.Buffer
	var payload GetBlocks

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}

	blocks := c.GetBlockHashes()
	SendInventory(payload.AddressFrom, "block", blocks)
}

func HandleGetData(request []byte, c *blockchain.Chain) {
	var buffer bytes.Buffer
	var payload GetData

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}

	if payload.Type == "block" {
		block, err := c.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		SendBlock(payload.AddressFrom, &block)
	}
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]
		SendTransaction(payload.AddressFrom, &tx)
	}
}

func HandleInventory(request []byte, c *blockchain.Chain) {
	var buffer bytes.Buffer
	var payload Inventory

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		SendGetData(payload.AddressFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}
	if payload.Type == "tx" {
		txID := payload.Items[0]

		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddressFrom, "tx", txID)
		}
	}
}

func HandleTransaction(request []byte, c *blockchain.Chain) {
	var buffer bytes.Buffer
	var payload Transaction

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}
	transactionData := payload.Transaction
	transaction := blockchain.DeserializeTransaction(transactionData)
	memoryPool[hex.EncodeToString(transaction.ID)] = transaction
	fmt.Printf("%s, %d", nodeAddress, len(memoryPool))
	if nodeAddress == KnownNodes[0] {
		//TODO is central node
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddressFrom {
				SendInventory(node, "tx", [][]byte{transaction.ID})
			}
		}
	} else {
		// TODO mine if something
		if len(memoryPool) >= 1 && len(minerAddress) > 0 {
			MineTx(c)
		}
	}
}

func MineTx(c *blockchain.Chain) {
	var txs []*blockchain.Transaction

	for id := range memoryPool {
		fmt.Printf("tx: %s\n", memoryPool[id].ID)
		tx := memoryPool[id]
		if c.VerifyTransaction(&tx) {
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		println("All transactions invalid")
		return
	}

	cbTx := blockchain.CoinbaseTransaction(minerAddress, "")
	txs = append(txs, cbTx)

	newBlock := c.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{Chain: c}
	UTXOSet.Reindex()

	fmt.Println("New block mined")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range KnownNodes {
		if node != nodeAddress {
			SendInventory(node, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(memoryPool) > 0 {
		MineTx(c)
	}
}

func HandleVersion(request []byte, c *blockchain.Chain) {
	var buffer bytes.Buffer
	var payload Version

	buffer.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err)
	}
	topHeight := c.GetTopHeight()
	otherHeight := payload.TopHeight

	if topHeight < otherHeight {
		SendGetBlocks(payload.AddressFrom)
	} else if topHeight > otherHeight {
		SendVersion(payload.AddressFrom, c)
	}

	if !NodeIsKnown(payload.AddressFrom) {
		KnownNodes = append(KnownNodes, payload.AddressFrom)
	}
}

func HandleConnection(conn net.Conn, chain *blockchain.Chain) {
	req, err := io.ReadAll(conn)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	command := BytesToCmd(req[:commandLength])
	fmt.Printf("Received %s command \n", command)

	switch command {
	case "adr":
		HandleAddress(req)
	case "blk":
		HandleBlock(req, chain)
	case "inv":
		HandleInventory(req, chain)
	case "gbk":
		HandleGetBlocks(req, chain)
	case "gdt":
		HandleGetData(req, chain)
	case "tnx":
		HandleTransaction(req, chain)
	case "vsn":
		HandleVersion(req, chain)
	default:
		println("Unknown command")
	}
}

func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

func RequestBlocks() {
	for _, node := range KnownNodes {
		SendGetBlocks(node)
	}
}

func NodeIsKnown(address string) bool {
	for _, node := range KnownNodes {
		if node == address {
			return true
		}
	}
	return false
}

func CmdToBytes(cmd string) []byte {
	var bytes [commandLength]byte
	for i, c := range cmd {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func BytesToCmd(bytes []byte) string {
	var cmd []byte
	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}
	return fmt.Sprintf("%s", cmd)
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)

	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	chain := blockchain.ContinueChain(nodeID)
	defer chain.Database.Close()
	go CloseDB(chain)

	if nodeAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], chain)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)
	}
}

func CloseDB(chain *blockchain.Chain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}
