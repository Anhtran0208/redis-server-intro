package core

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Anhtran0208/redis-server-intro/internal/constant"
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

type Executor struct {
	store *Store
}

func NewExecutor(store *Store) *Executor {
	return &Executor{
		store: store,
	}
}

func cmdPing(args []string) []byte {
	var res []byte
	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of argument for 'ping' command"), false)
	}
	// simulation
	//time.Sleep(20 * time.Second)
	if len(args) == 0 {
		res = Encode("PONG", true)
	} else {
		res = Encode(args[0], false)
	}
	return res
}

func (e *Executor) cmdSet(args []string) []byte {
	if len(args) < 2 || len(args) == 3 || len(args) > 4 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'SET' command"), false)
	}

	var key, value string
	var ttlMs int64 = -1

	key, value = args[0], args[1]
	if len(args) > 2 {
		ttlSec, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
		}
		ttlMs = ttlSec * 1000
	}
	e.store.Dict.Set(key, e.store.Dict.NewObj(key, value, ttlMs))
	return constant.RespOk
}

func (e *Executor) cmdGet(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'GET' command"), false)
	}
	key := args[0]
	obj := e.store.Dict.Get(key)
	if obj == nil {
		return constant.RespNil
	}

	if e.store.Dict.HasExpired(key) {
		return constant.RespNil
	}

	return Encode(obj.Value, false)
}

func (e *Executor) cmdTTL(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'TTL' command"), false)
	}
	key := args[0]
	obj := e.store.Dict.Get(key)
	if obj == nil {
		return constant.TtlKeyNotExist
	}

	exp, isExpirySet := e.store.Dict.GetExpiry(key)
	if !isExpirySet {
		return constant.TtlKeyExistNoExpire
	}

	remainMs := exp - uint64(time.Now().UnixMilli())
	if remainMs < 0 {
		return constant.TtlKeyNotExist
	}

	return Encode(int64(remainMs/1000), false)
}

func cmdINFO(args []string) []byte {
	var info []byte
	buff := bytes.NewBuffer(info)
	buff.WriteString("# Keyspace\r\n")
	buff.WriteString(fmt.Sprintf("db0:keys=%d,expires=0,avg_ttl=0\r\n", data_structure.HashKeySpaceStat.Key))
	return Encode(buff.String(), false)
}

func (e *Executor) ExecuteCMD(cmd *Command) []byte {
	switch cmd.Cmd {
	// ping, get, set, ttl
	case "PING":
		return cmdPing(cmd.Args)
	case "GET":
		return e.cmdGet(cmd.Args)
	case "SET":
		return e.cmdSet(cmd.Args)
	case "TTL":
		return e.cmdTTL(cmd.Args)

	// simple set
	case "SADD":
		return e.cmdSADD(cmd.Args)
	case "SREM":
		return e.cmdSREM(cmd.Args)
	case "SMEMBERS":
		return e.cmdSMEMBERS(cmd.Args)
	case "SISMEMBER":
		return e.cmdSISMEMBER(cmd.Args)

	// sorted set
	case "ZADD":
		return e.cmdZADD(cmd.Args)
	case "ZSCORE":
		return e.cmdZSCORE(cmd.Args)
	case "ZRANK":
		return e.cmdZRANK(cmd.Args)

	// counter min sketch
	case "CMS.INITBYDIM":
		return e.cmdCMSINITBYDIM(cmd.Args)
	case "CMS.INITBYPROB":
		return e.cmdCMSINITBYPROB(cmd.Args)
	case "CMS.INCRBY":
		return e.cmdCMSINCRBY(cmd.Args)
	case "CMS.QUERY":
		return e.cmdCMSQUERY(cmd.Args)

	// bloom filter
	case "BF.RESERVE":
		return e.cmdBFRESERVE(cmd.Args)
	case "BF.MADD":
		return e.cmdBFMADD(cmd.Args)
	case "BF.EXISTS":
		return e.cmdBFEXISTS(cmd.Args)

	// info cmd
	case "INFO":
		return cmdINFO(cmd.Args)
	default:
		return []byte("-CMD not found\r\n")
	}
}
