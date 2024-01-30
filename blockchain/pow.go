package blockchain

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"runtime"
	"sync"

	"github.com/JI-0/private-cryptocurrency/randomx"
)

const difficulty = 4
const largePages = false
const initKey = "bc2bcbb0f927bac40faaf98a468f4de5e81b9395ba6c970634abb4d7b1cb007b"

var once sync.Once
var flagsCache randomx.Flag

var activeKeyNum = 0
var activeKey = []byte("bc2bcbb0f927bac40faaf98a468f4de5e81b9395ba6c970634abb4d7b1cb007b")

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
	cache  randomx.Cache
	ds     randomx.Dataset
	VM     randomx.VM
}

func NewProof(c *Chain, b *Block, fullMem bool) *ProofOfWork {
	once.Do(func() {
		flagsCache = randomx.GetFlags()
	})
	flags := flagsCache
	if fullMem {
		flags |= randomx.FlagFullMEM
	}
	if largePages {
		flags |= randomx.FlagLargePages
	}

	reqKeyNum := b.Height / 2048
	if b.Height%2048 >= 64 {
		reqKeyNum++
	}
	if reqKeyNum != activeKeyNum {
		activeKeyNum = reqKeyNum
		if reqKeyNum == 0 {
			activeKey = []byte(initKey)
		} else {
			targetHeight := 2048 * (reqKeyNum - 1)
			iter := c.Iterator()
			for {
				bck := iter.Next()
				if bck.Height == targetHeight {
					activeKey = bck.Hash
					break
				}
			}
		}
	}

	cache, err := randomx.AllocCache(flags)
	if err != nil {
		println(err)
	}
	randomx.InitCache(cache, activeKey)
	ds, err := randomx.AllocDataset(flags)
	if err != nil {
		println(err)
	}

	count := randomx.DatasetItemCount()
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		go func() {
			defer wg.Done()
			randomx.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()

	vm, err := randomx.CreateVM(cache, ds, flags)
	if err != nil {
		println(err)
	}

	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))
	pow := &ProofOfWork{b, target, cache, ds, vm}
	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(difficulty)),
		}, []byte{})
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash []byte
	nonce := 0

	// Init mining
	data := pow.InitData(nonce)
	randomx.CalculateHashFirst(pow.VM, data)
	nonce++

	// Mine
	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = randomx.CalculateHashNext(pow.VM, data)
		fmt.Printf("\r%x", hash)

		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	nonce--

	fmt.Println()

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)
	hash := randomx.CalculateHash(pow.VM, data)

	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

func (pow *ProofOfWork) Destroy() {
	randomx.DestroyVM(pow.VM)
	randomx.ReleaseDataset(pow.ds)
	randomx.ReleaseCache(pow.cache)
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}
