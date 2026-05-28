package core

import (
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

type Store struct {
	Dict  *data_structure.Dict
	Set   map[string]*data_structure.SimpleSet
	ZSet  map[string]*data_structure.SortedSet
	CMS   map[string]*data_structure.CMS
	Bloom map[string]*data_structure.BloomFilter
}

func NewStore(dictConfig data_structure.DictConfig) *Store {
	return &Store{
		Dict:  data_structure.CreateDict(dictConfig),
		Set:   make(map[string]*data_structure.SimpleSet),
		ZSet:  make(map[string]*data_structure.SortedSet),
		CMS:   make(map[string]*data_structure.CMS),
		Bloom: make(map[string]*data_structure.BloomFilter),
	}
}

func (s *Store) ActiveDeleteExpiredKeys() {
	s.Dict.ActiveDeleteExpiredKeys()
}
