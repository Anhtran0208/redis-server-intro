package core

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Anhtran0208/redis-server-intro/internal/constant"
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

func (e *Executor) cmdBFRESERVE(args []string) []byte {
	if !(len(args) == 3 || len(args) == 5) {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'BF.RESERVE' command"), false)
	}
	key := args[0]
	errRate, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return Encode(errors.New(fmt.Sprintf("error rate must be a floating point number %s", args[1])), false)
	}
	capacity, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return Encode(errors.New(fmt.Sprintf("capacity must be an integer number %s", args[2])), false)
	}
	_, exist := e.store.Bloom[key]
	if exist {
		return Encode(errors.New(fmt.Sprintf("Bloom filter with key '%s' already exist", key)), false)
	}
	e.store.Bloom[key] = data_structure.CreateBloomFilter(capacity, errRate)
	return constant.RespOk
}

func (e *Executor) cmdBFMADD(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'BF.MADD' command"), false)
	}
	key := args[0]
	bloom, exist := e.store.Bloom[key]
	if !exist {
		bloom = data_structure.CreateBloomFilter(constant.BfDefaultInitCapacity,
			constant.BfDefaultErrRate)
		e.store.Bloom[key] = bloom
	}
	var res []string
	for i := 1; i < len(args); i++ {
		item := args[i]
		bloom.Add(item)
		res = append(res, "1")
	}
	return Encode(res, false)
}

func (e *Executor) cmdBFEXISTS(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'BF.EXISTS' command"), false)
	}
	key, item := args[0], args[1]
	bloom, exist := e.store.Bloom[key]
	if !exist {
		return constant.RespZero
	}
	if !bloom.Exist(item) {
		return constant.RespZero
	}
	return constant.RespOne
}
