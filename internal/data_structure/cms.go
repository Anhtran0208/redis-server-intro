package data_structure

import (
	"math"

	"github.com/twmb/murmur3"
)

// Log10PointFive is a precomputed value for log10(0.5).
const Log10PointFive = -0.30102999566

type CMS struct {
	rows    uint32
	cols    uint32
	counter [][]uint32
}

func CreateCMS(rows, cols uint32) *CMS {
	cms := &CMS{
		rows: rows,
		cols: cols,
	}
	cms.counter = make([][]uint32, rows)
	for i := uint32(0); i < rows; i++ {
		cms.counter[i] = make([]uint32, cols)
	}
	return cms
}

// calc hash function using murmur3 library
func (cms *CMS) calcHash(item string, rowIdx uint32) uint32 {
	hash := murmur3.SeedNew32(rowIdx)
	hash.Write([]byte(item))
	return hash.Sum32()
}

// increase x by value
func (cms *CMS) Increase(item string, value uint32) uint32 {
	var minCnt uint32 = math.MaxUint32

	// loop through each row of 2d array
	for i := uint32(0); i < cms.rows; i++ {
		// calc hash function
		hash := cms.calcHash(item, i)
		col := hash % cms.cols

		// prevent overflow
		if math.MaxUint32-cms.counter[i][col] < value {
			cms.counter[i][col] = math.MaxUint32
		} else {
			cms.counter[i][col] += value
		}

		// keep track the min cnt to get the freq
		if cms.counter[i][col] < minCnt {
			minCnt = cms.counter[i][col]
		}
	}
	return minCnt
}

// estimate cnt for an item
func (cms *CMS) Count(item string) uint32 {
	var minCnt uint32 = math.MaxUint32

	for i := uint32(0); i < cms.rows; i++ {
		// calc hash function
		hash := cms.calcHash(item, i)
		col := hash % cms.cols

		if cms.counter[i][col] < minCnt {
			minCnt = cms.counter[i][col]
		}
	}
	return minCnt
}

// CalcCMSDim calculates the dimensions (cols and row) of the CMS
// based on the desired error rate and probability.
func CalcCMSDim(errRate float64, errProb float64) (uint32, uint32) {
	cols := uint32(math.Ceil(2.0 / errRate))
	rows := uint32(math.Ceil(math.Log10(errProb) / Log10PointFive))
	return cols, rows
}
