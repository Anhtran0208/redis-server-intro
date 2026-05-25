package data_structure

import (
	"math"

	"github.com/spaolacci/murmur3"
)

type BloomFilter struct {
	ErrorRate       float64
	NumberOfEntries uint64
	NumberOfHash    int
	BitArr          []uint8
	bitPerEntry     float64
	NumberOfBits    uint64 // size of bit arr in bit
	NumberOfBytes   uint64 // size of bit arr in byte
}

const Ln2 float64 = 0.693147180559945
const Ln2Square float64 = 0.480453013918201
const ABigSeed uint32 = 0x9747b28c

type HashValue struct {
	a uint64
	b uint64
}

/*
http://en.wikipedia.org/wiki/Bloom_filter
- Optimal number of bits is: bits = (entries * ln(error)) / ln(2)^2
- bitPerEntry = bits/entries
- Optimal number of hash functions is: hashes = bitPerEntry * ln(2)
*/
func calcBitsPerEntry(errRate float64) float64 {
	num := math.Log(errRate)
	return math.Abs(-(num / Ln2Square))
}

func CreateBloomFilter(numberOfEntries uint64, errorRate float64) *BloomFilter {
	bloomFilter := BloomFilter{
		NumberOfEntries: numberOfEntries,
		ErrorRate:       errorRate,
	}

	// bit per entry
	bloomFilter.bitPerEntry = calcBitsPerEntry(errorRate)

	// number of bits
	numberOfBits := uint64(float64(numberOfEntries) * bloomFilter.bitPerEntry)

	if numberOfBits%64 != 0 {
		bloomFilter.NumberOfBytes = ((numberOfBits / 64) + 1) * 8
	} else {
		bloomFilter.NumberOfBytes = numberOfBits / 8
	}

	bloomFilter.NumberOfBits = bloomFilter.NumberOfBytes * 8

	// number of hash
	bloomFilter.NumberOfHash = int(math.Ceil(Ln2 * bloomFilter.bitPerEntry))
	// bit arr
	bloomFilter.BitArr = make([]uint8, bloomFilter.NumberOfBytes)
	return &bloomFilter
}

// double hashing tech
// generate 2 hash value

func (bloomFilter *BloomFilter) GenerateHashValue(entry string) HashValue {
	hasher := murmur3.New128WithSeed(ABigSeed)

	hasher.Write([]byte(entry))
	x, y := hasher.Sum128()
	return HashValue{
		a: x,
		b: y,
	}
}

func (bloomFilter *BloomFilter) Add(entry string) {
	var hash, bytePos uint64
	initHash := bloomFilter.GenerateHashValue(entry)
	for i := 0; i < bloomFilter.NumberOfHash; i++ {
		hash = (initHash.a + initHash.b*uint64(i)) % bloomFilter.NumberOfBits
		bytePos = hash >> 3 // div 8

		// flip bit
		bloomFilter.BitArr[bytePos] |= 1 << (hash % 8)
	}
}

func (bloomFilter *BloomFilter) Exist(entry string) bool {
	var hash, bytePos uint64
	initHash := bloomFilter.GenerateHashValue(entry)
	for i := 0; i < bloomFilter.NumberOfHash; i++ {
		hash = (initHash.a + initHash.b*uint64(i)) % bloomFilter.NumberOfBits
		bytePos = hash >> 3 // div 8

		// if any bit = 0 => return false
		if (bloomFilter.BitArr[bytePos] & (1 << (hash % 8))) == 0 {
			return false
		}
	}
	return true
}
