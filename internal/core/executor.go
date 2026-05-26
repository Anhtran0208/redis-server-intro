package core

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"syscall"
	"time"

	"github.com/Anhtran0208/redis-server-intro/internal/constant"
	"github.com/Anhtran0208/redis-server-intro/internal/data_structure"
)

func cmdPing(args []string) []byte {
	var res []byte
	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of argument for 'ping' command"), false)
	}
	// simulation
	time.Sleep(20 * time.Second)
	if len(args) == 0 {
		res = Encode("PONG", true)
	} else {
		res = Encode(args[0], false)
	}
	return res
}

func cmdSet(args []string) []byte {
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
	dictStore.Set(key, dictStore.NewObj(key, value, ttlMs))
	return constant.RespOk
}

func cmdGet(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'GET' command"), false)
	}
	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return constant.RespNil
	}

	if dictStore.HasExpired(key) {
		return constant.RespNil
	}

	return Encode(obj.Value, false)
}

func cmdTTL(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'TTL' command"), false)
	}
	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return constant.TtlKeyNotExist
	}

	exp, isExpirySet := dictStore.GetExpiry(key)
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

func ExecuteAndResponse(cmd *Command, connFd int) error {
	var res []byte
	switch cmd.Cmd {
	// ping, get, set, ttl
	case "PING":
		res = cmdPing(cmd.Args)
	case "GET":
		res = cmdGet(cmd.Args)
	case "SET":
		res = cmdSet(cmd.Args)
	case "TTL":
		res = cmdTTL(cmd.Args)

	// simple set
	case "SADD":
		res = cmdSADD(cmd.Args)
	case "SREM":
		res = cmdSREM(cmd.Args)
	case "SMEMBERS":
		res = cmdSMEMBERS(cmd.Args)
	case "SISMEMBER":
		res = cmdSISMEMBER(cmd.Args)

	// sorted set
	case "ZADD":
		res = cmdZADD(cmd.Args)
	case "ZSCORE":
		res = cmdZSCORE(cmd.Args)
	case "ZRANK":
		res = cmdZRANK(cmd.Args)

	// counter min sketch
	case "CMS.INITBYDIM":
		res = cmdCMSINITBYDIM(cmd.Args)
	case "CMS.INITBYPROB":
		res = cmdCMSINITBYPROB(cmd.Args)
	case "CMS.INCRBY":
		res = cmdCMSINCRBY(cmd.Args)
	case "CMS.QUERY":
		res = cmdCMSQUERY(cmd.Args)

	// bloom filter
	case "BF.RESERVE":
		res = cmdBFRESERVE(cmd.Args)
	case "BF.MADD":
		res = cmdBFMADD(cmd.Args)
	case "BF.EXISTS":
		res = cmdBFEXISTS(cmd.Args)

	// info cmd
	case "INFO":
		res = cmdINFO(cmd.Args)
	default:
		res = []byte(fmt.Sprintf("-CMD not found\r\n"))
	}
	_, err := syscall.Write(connFd, res)
	return err
}
