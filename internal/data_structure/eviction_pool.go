package data_structure

import (
	"sort"
)

type EpoolConfig struct {
	EpoolMaxSize int
}
type EvictionCandidate struct {
	key            string
	lastAccessTime uint32
}

type EvictionPool struct {
	evictionPool []*EvictionCandidate
	cfg          EpoolConfig
}

type ByLastAccessTime []*EvictionCandidate

func (arr ByLastAccessTime) Len() int {
	return len(arr)
}

func (arr ByLastAccessTime) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

func (arr ByLastAccessTime) Less(i, j int) bool {
	return arr[i].lastAccessTime < arr[j].lastAccessTime
}

// push new item to pool, maintain lastAccessTime accending order
// if pool size > max_size, remove newest item
func (pool *EvictionPool) Push(key string, lastAccessTime uint32) {
	newItem := &EvictionCandidate{
		key:            key,
		lastAccessTime: lastAccessTime,
	}

	// check if item is exist or not
	exist := false
	for i := 0; i < len(pool.evictionPool); i++ {
		if pool.evictionPool[i].key == key {
			exist = true
			pool.evictionPool[i] = newItem
		}
	}
	if !exist {
		pool.evictionPool = append(pool.evictionPool, newItem)
	}
	// sort pool by last access time
	sort.Sort(ByLastAccessTime(pool.evictionPool))

	// remove newest item if pool is full
	if len(pool.evictionPool) > pool.cfg.EpoolMaxSize {
		lastIdx := len(pool.evictionPool) - 1
		key = pool.evictionPool[lastIdx].key
		pool.evictionPool = pool.evictionPool[:lastIdx]
	}
}

// pop the oldest item from the pool
func (pool *EvictionPool) Pop() *EvictionCandidate {
	if len(pool.evictionPool) == 0 {
		return nil
	}
	oldestItem := pool.evictionPool[0]
	pool.evictionPool = pool.evictionPool[1:]
	return oldestItem
}

func NewEvictionPool(cfg EpoolConfig) *EvictionPool {
	return &EvictionPool{
		evictionPool: make([]*EvictionCandidate, 0),
		cfg:          cfg,
	}
}
