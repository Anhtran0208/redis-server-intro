package data_structure

import (
	"log"
	"time"

	"github.com/Anhtran0208/redis-server-intro/internal/config"
)

type Obj struct {
	Value          interface{}
	LastAccessTime uint32
}

type Dict struct {
	// key => value
	dictStore map[string]*Obj

	// key => expiration time
	expiredDictStore map[string]uint64
}

func CreateDict() *Dict {
	res := Dict{
		dictStore:        make(map[string]*Obj),
		expiredDictStore: make(map[string]uint64),
	}
	return &res
}

func (d *Dict) GetExpireDictStore() map[string]uint64 {
	return d.expiredDictStore
}

func (d *Dict) GetDictStore() map[string]*Obj {
	return d.dictStore
}

func now() uint32 {
	return uint32(time.Now().Unix())
}

func (d *Dict) NewObj(key string, value interface{}, ttlMs int64) *Obj {
	obj := &Obj{
		Value:          value,
		LastAccessTime: now(),
	}
	if ttlMs > 0 {
		d.SetExpiry(key, ttlMs)
	}
	return obj
}

func (d *Dict) SetExpiry(key string, ttlMs int64) {
	d.expiredDictStore[key] = uint64(time.Now().UnixMilli()) + uint64(ttlMs)
}

func (d *Dict) GetExpiry(key string) (uint64, bool) {
	exp, exist := d.expiredDictStore[key]
	return exp, exist
}

func (d *Dict) HasExpired(key string) bool {
	exp, exist := d.expiredDictStore[key]
	if !exist {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

func (d *Dict) Get(key string) *Obj {
	value := d.dictStore[key]
	if value != nil {
		value.LastAccessTime = now()
		if d.HasExpired(key) {
			d.Delete(key)
			return nil
		}
	}
	return value
}

func (d *Dict) Set(key string, obj *Obj) {
	if len(d.dictStore) == config.MaxKeyNumber {
		d.evict()
	}

	value := d.dictStore[key]
	if value == nil {
		HashKeySpaceStat.Key++
	}
	d.dictStore[key] = obj
}
func (d *Dict) Delete(key string) bool {
	if _, exist := d.dictStore[key]; exist {
		log.Printf("Delete key %s", key)
		delete(d.dictStore, key)
		delete(d.expiredDictStore, key)

		HashKeySpaceStat.Key--
		return true
	}
	return false
}

// evict random
func (d *Dict) evictRandom() {
	evictCnt := int64(config.EvictionRatio * float64(config.MaxKeyNumber))
	log.Print("trigger random eviction")
	for key := range d.dictStore {
		d.Delete(key)
		evictCnt--
		if evictCnt == 0 {
			break
		}
	}
}
func (d *Dict) evict() {
	switch config.EvictionPolicy {
	case "allkeys-random":
		d.evictRandom()
	case "allkeys-lru":
		d.evictLRU()
	}
}

// sample 5 random keys
func (d *Dict) populateEpool() {
	remainSampleSize := config.EpoolLruSampleSize
	for key := range d.dictStore {
		ePool.Push(key, d.dictStore[key].LastAccessTime)
		remainSampleSize--
		if remainSampleSize == 0 {
			break
		}
	}
	log.Println("Epool:")
	for _, item := range ePool.evictionPool {
		log.Println(item.key, item.lastAccessTime)
	}
}

// evic approximate lru
func (d *Dict) evictLRU() {
	d.populateEpool()
	evictCount := int64(config.EvictionRatio * float64(config.MaxKeyNumber))
	log.Print("trigger LRU eviction")
	for i := 0; i < int(evictCount) && len(ePool.evictionPool) > 0; i++ {
		item := ePool.Pop()
		if item != nil {
			d.Delete(item.key)
		}
	}
}
