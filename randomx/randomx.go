package randomx

//#cgo CFLAGS: -I./randomx
//#cgo LDFLAGS: -L${SRCDIR}/lib -lrandomx -lstdc++ -lm -lpthread
//#include <stdlib.h>
//#include "RandomX/src/randomx.h"
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

const RxHashSize = C.RANDOMX_HASH_SIZE

type Flag int

// All flags
const (
	FlagDefault     Flag = 0 // for all default
	FlagLargePages  Flag = 1 // for dataset & rxCache & vm
	FlagHardAES     Flag = 2 // for vm
	FlagFullMEM     Flag = 4 // for vm
	FlagJIT         Flag = 8 // for vm & cache
	FlagSecure      Flag = 16
	FlagArgon2SSSE3 Flag = 32 // for cache
	FlagArgon2AVX2  Flag = 64 // for cache
	FlagArgon2      Flag = 96 // = avx2 + sse3
)

func (f Flag) toC() C.randomx_flags {
	return C.randomx_flags(f)
}

type Cache *C.randomx_cache

type Dataset *C.randomx_dataset

type VM *C.randomx_vm

func GetFlags() Flag {
	flag := Flag(C.randomx_get_flags())
	if runtime.GOARCH == "arm64" && runtime.GOOS == "darwin" {
		flag = flag | FlagSecure
	}
	return flag
}

func AllocCache(flags ...Flag) (Cache, error) {
	var SumFlag = FlagDefault
	var cache *C.randomx_cache

	for _, flag := range flags {
		SumFlag = SumFlag | flag
	}

	cache = C.randomx_alloc_cache(SumFlag.toC())
	if cache == nil {
		return nil, errors.New("failed to alloc mem for rxCache")
	}

	return cache, nil
}

func InitCache(cache Cache, seed []byte) {
	if len(seed) == 0 {
		panic("seed cannot be NULL")
	}

	C.randomx_init_cache(cache, unsafe.Pointer(&seed[0]), C.size_t(len(seed)))
}

func ReleaseCache(cache Cache) {
	C.randomx_release_cache(cache)
}

func AllocDataset(flags ...Flag) (Dataset, error) {
	var SumFlag = FlagDefault
	for _, flag := range flags {
		SumFlag = SumFlag | flag
	}

	var dataset *C.randomx_dataset
	dataset = C.randomx_alloc_dataset(SumFlag.toC())
	if dataset == nil {
		return nil, errors.New("failed to alloc mem for dataset")
	}

	return dataset, nil
}

func DatasetItemCount() uint32 {
	var length C.ulong
	length = C.randomx_dataset_item_count()
	return uint32(length)
}

func InitDataset(dataset Dataset, cache Cache, startItem uint32, itemCount uint32) {
	if dataset == nil {
		panic("alloc dataset mem is required")
	}

	if cache == nil {
		panic("alloc cache mem is required")
	}

	C.randomx_init_dataset(dataset, cache, C.ulong(startItem), C.ulong(itemCount))
}

func GetDatasetMemory(dataset Dataset) unsafe.Pointer {
	return C.randomx_get_dataset_memory(dataset)
}

func ReleaseDataset(dataset Dataset) {
	C.randomx_release_dataset(dataset)
}

func CreateVM(cache Cache, dataset Dataset, flags ...Flag) (VM, error) {
	var SumFlag = FlagDefault
	for _, flag := range flags {
		SumFlag = SumFlag | flag

		// Dataset may not be nil if FullMem flag is set
		if flag == FlagFullMEM && dataset == nil {
			return nil, errors.New("FlagFullMEM requires initialized dataset")
		}
	}

	vm := C.randomx_create_vm(SumFlag.toC(), cache, dataset)

	if vm == nil {
		return nil, errors.New("failed to create vm")
	}

	return vm, nil
}

func SetVMCache(vm VM, cache Cache) {
	C.randomx_vm_set_cache(vm, cache)
}

func SetVMDataset(vm VM, dataset Dataset) {
	C.randomx_vm_set_dataset(vm, dataset)
}

func DestroyVM(vm VM) {
	C.randomx_destroy_vm(vm)
}

func CalculateHash(vm VM, in []byte) []byte {
	if vm == nil {
		panic("failed hashing: using empty vm")
	}

	input := C.CBytes(in)
	output := C.CBytes(make([]byte, RxHashSize))
	C.randomx_calculate_hash(vm, input, C.size_t(len(in)), output)
	hash := C.GoBytes(output, RxHashSize)
	C.free(unsafe.Pointer(input))
	C.free(unsafe.Pointer(output))

	return hash
}

func CalculateHashFirst(vm VM, in []byte) {
	if vm == nil {
		panic("failed hashing: using empty vm")
	}

	input := C.CBytes(in)
	C.randomx_calculate_hash_first(vm, input, C.size_t(len(in)))
	C.free(unsafe.Pointer(input))
}

func CalculateHashNext(vm VM, in []byte) []byte {
	if vm == nil {
		panic("failed hashing: using empty vm")
	}

	input := C.CBytes(in)
	output := C.CBytes(make([]byte, RxHashSize))
	C.randomx_calculate_hash_next(vm, input, C.size_t(len(in)), output)
	hash := C.GoBytes(output, RxHashSize)
	C.free(unsafe.Pointer(input))
	C.free(unsafe.Pointer(output))

	return hash
}
