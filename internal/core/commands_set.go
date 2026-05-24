package core

import (
	"errors"

	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

// SADD cmd: add members to set
func cmdSADD(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SADD' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		set = data_structure.NewSimpleSet(key)
		setStore[key] = set
	}

	cntAdded := set.Add(args[1:]...)
	return Encode(cntAdded, false)
}

// SREM cmd: remove members to set
func cmdSREM(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SREM' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		set = data_structure.NewSimpleSet(key)
		setStore[key] = set
	}

	cntRemoved := set.Remove(args[1:]...)
	return Encode(cntRemoved, false)
}

// SMEMBERS cmd => list all members in set
func cmdSMEMBERS(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SMEMBERS' command"), false)
	}
	key := args[0]
	set, exist := setStore[key]
	if !exist {
		return Encode(make([]string, 0), false)
	}
	return Encode(set.ListMembers(), false)
}

// SISMEMBER cmd => check if member is in the set
func cmdSISMEMBER(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SISMEMBER' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		return Encode(0, false)
	}
	return Encode(set.IsMember(args[1]), false)
}
