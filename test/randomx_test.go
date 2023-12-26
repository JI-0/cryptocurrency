package test

import (
	"bytes"
	"encoding/hex"
	"runtime"
	"sync"
	"testing"

	"github.com/JI-0/private-cryptocurrency/randomx"
)

var testPairs = [][][]byte{
	// randomX
	{
		[]byte("test key 000"),
		[]byte("This is a test"),
		[]byte("639183aae1bf4c9a35884cb46b09cad9175f04efd7684e7262a0ac1c2f0b4e3f"), //"c90f60ba2ebed5c7404e035853d90c3f8416b303125ca2e05dee529bb35df627"),
	},
}

func TestAllocCache(t *testing.T) {
	cache, _ := randomx.AllocCache(randomx.FlagDefault)
	randomx.InitCache(cache, []byte("123"))
	randomx.ReleaseCache(cache)
}

func TestAllocDataset(t *testing.T) {
	t.Log("warning: cannot use FlagDefault only, very slow!. After using FlagJIT, really fast!")

	ds, err := randomx.AllocDataset(randomx.FlagJIT)
	if err != nil {
		panic(err)
	}
	cache, err := randomx.AllocCache(randomx.FlagJIT)
	if err != nil {
		panic(err)
	}

	seed := make([]byte, 32)
	randomx.InitCache(cache, seed)
	t.Log("rxCache initialization finished")

	count := randomx.DatasetItemCount()
	t.Log("dataset count:", count/1024/1024, "mb")
	randomx.InitDataset(ds, cache, 0, count)
	t.Log(randomx.GetDatasetMemory(ds))

	randomx.ReleaseDataset(ds)
	randomx.ReleaseCache(cache)
}

func TestCreateVM(t *testing.T) {
	var tp = testPairs[0]

	t.Log("Allocate cache")
	cache, _ := randomx.AllocCache(randomx.FlagDefault)

	t.Log("Initialize cache")
	seed := tp[0]
	randomx.InitCache(cache, seed)

	t.Log("Allocate dataset")
	ds, _ := randomx.AllocDataset(randomx.FlagDefault)

	t.Log("Initialize dataset")
	count := randomx.DatasetItemCount()
	t.Log("  -- dataset item count:", count)
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	t.Logf("  -- initializing with %d workers", workerNum)
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

	t.Log("Create VM")
	vm, err := randomx.CreateVM(cache, ds, randomx.FlagDefault)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	var hashCorrect = make([]byte, hex.DecodedLen(len(tp[2])))
	_, err = hex.Decode(hashCorrect, tp[2])
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("Compute hash and check result")
	hash := randomx.CalculateHash(vm, tp[1])
	if !bytes.Equal(hash, hashCorrect) {
		t.Logf("answer is incorrect: %x, %x", hash, hashCorrect)
		t.Fail()
	}
}

func TestCreateRecommendedFlagVM(t *testing.T) {
	flags := randomx.GetFlags() //randomx.GetFlags() //| randomx.FlagFullMEM
	t.Log("Flags: ", flags)

	var tp = testPairs[0]

	t.Log("Allocate cache")
	cache, _ := randomx.AllocCache(flags)

	t.Log("Initialize cache")
	seed := tp[0]
	randomx.InitCache(cache, seed)

	t.Log("Allocate dataset")
	ds, _ := randomx.AllocDataset(flags)

	t.Log("Initialize dataset")
	count := randomx.DatasetItemCount()
	t.Log("  -- dataset item count:", count)
	var wg sync.WaitGroup
	var workerNum = uint32(runtime.NumCPU())
	t.Logf("  -- initializing with %d workers", workerNum)
	for i := uint32(0); i < workerNum; i++ {
		wg.Add(1)
		a := (count * i) / workerNum
		b := (count * (i + 1)) / workerNum
		t.Log("  -- worker initialized:", i, a, b-a)
		go func() {
			defer wg.Done()
			randomx.InitDataset(ds, cache, a, b-a)
		}()
	}
	wg.Wait()

	t.Log("Create VM")
	vm, err := randomx.CreateVM(cache, ds, flags)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	var hashCorrect = make([]byte, hex.DecodedLen(len(tp[2])))
	_, err = hex.Decode(hashCorrect, tp[2])
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("Compute hash and check result")
	hash := randomx.CalculateHash(vm, tp[1])
	if !bytes.Equal(hash, hashCorrect) {
		t.Logf("answer is incorrect: %x, %x", hash, hashCorrect)
		t.Fail()
	}
}
