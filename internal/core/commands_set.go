package core

import (
	"errors"

	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

// SADD cmd: add members to set
func (e *Executor) cmdSADD(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SADD' command"), false)
	}

	key := args[0]
	set, exist := e.store.Set[key]
	if !exist {
		set = data_structure.NewSimpleSet(key)
		e.store.Set[key] = set
	}

	cntAdded := set.Add(args[1:]...)
	return Encode(cntAdded, false)
}

// SREM cmd: remove members to set
func (e *Executor) cmdSREM(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SREM' command"), false)
	}

	key := args[0]
	set, exist := e.store.Set[key]
	if !exist {
		set = data_structure.NewSimpleSet(key)
		e.store.Set[key] = set
	}

	cntRemoved := set.Remove(args[1:]...)
	return Encode(cntRemoved, false)
}

// SMEMBERS cmd => list all members in set
func (e *Executor) cmdSMEMBERS(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SMEMBERS' command"), false)
	}
	key := args[0]
	set, exist := e.store.Set[key]
	if !exist {
		return Encode(make([]string, 0), false)
	}
	return Encode(set.ListMembers(), false)
}

// SISMEMBER cmd => check if member is in the set
func (e *Executor) cmdSISMEMBER(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SISMEMBER' command"), false)
	}

	key := args[0]
	set, exist := e.store.Set[key]
	if !exist {
		return Encode(0, false)
	}
	return Encode(set.IsMember(args[1]), false)
}
