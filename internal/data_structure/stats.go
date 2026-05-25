package data_structure

type KeySpaceStat struct {
	Key    int64
	Expire int64 //number of expired keys that hasn't been deleted
}

var HashKeySpaceStat = KeySpaceStat{
	Key:    0,
	Expire: 0,
}
