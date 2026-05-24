package data_structure

// set: name of set -> elements of the set
type SimpleSet struct {
	key  string
	dict map[string]struct{}
}

func NewSimpleSet(key string) *SimpleSet {
	return &SimpleSet{
		key:  key,
		dict: make(map[string]struct{}),
	}
}

// add operation -> return number of added elements
func (simpleSet *SimpleSet) Add(members ...string) int {
	cntAdded := 0
	for _, ele := range members {
		if _, exist := simpleSet.dict[ele]; !exist {
			simpleSet.dict[ele] = struct{}{}
			cntAdded++
		}
	}
	return cntAdded
}

// remove operation -> return number of removed elements
func (simpleSet *SimpleSet) Remove(members ...string) int {
	cntRemoved := 0
	for _, ele := range members {
		if _, exist := simpleSet.dict[ele]; exist {
			delete(simpleSet.dict, ele)
			cntRemoved++
		}
	}
	return cntRemoved
}

// check if element is in set
func (simpleSet *SimpleSet) IsMember(member string) int {
	_, exist := simpleSet.dict[member]
	if !exist {
		return 0
	}
	return 1
}

// list all members of set
func (simpleSet *SimpleSet) ListMembers() []string {
	allMembers := make([]string, 0, len(simpleSet.dict))
	for key, _ := range simpleSet.dict {
		allMembers = append(allMembers, key)
	}
	return allMembers
}
