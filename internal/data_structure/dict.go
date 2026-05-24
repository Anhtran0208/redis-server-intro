package data_structure

import (
	"time"
)

type Obj struct {
	Value interface{}
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
		Value: value,
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
		if d.HasExpired(key) {
			d.Delete(key)
			return nil
		}
	}
	return value
}

func (d *Dict) Set(key string, obj *Obj) {
	d.dictStore[key] = obj
}
func (d *Dict) Delete(key string) bool {
	if _, exist := d.dictStore[key]; exist {
		delete(d.dictStore, key)
		delete(d.expiredDictStore, key)
		return true
	}
	return false
}
