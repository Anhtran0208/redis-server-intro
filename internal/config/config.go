package config

import (
	"flag"
	"fmt"
)

type ExecutionMode string

const (
	SingleThreadMode ExecutionMode = "single-thread"
	MultiThreadMode  ExecutionMode = "multi-thread"
)

type Config struct {
	Protocol      string
	Port          string
	Mode          ExecutionMode
	NumWorkers    int
	NumIOHandlers int
	NumListeners  int

	MaxConnection int
	MaxKeyNumber  int

	EvictionPolicy     string
	EvictionRatio      float64
	EpoolMaxSize       int
	EpoolLruSampleSize int
}

func Load() (*Config, error) {
	mode := flag.String("mode", "multi", "execution mode: single or multi")
	port := flag.String("port", "3000", "server port")
	protocol := flag.String("protocol", "tcp", "network protocol")

	workers := flag.Int("workers", 4, "number of worker goroutines for multi-worker mode")
	ioHandlers := flag.Int("io-handlers", 4, "number of I/O handlers for multi-worker mode")
	listeners := flag.Int("listeners", 3, "number of socket listeners")

	maxConn := flag.Int("max-connections", 20000, "maximum number of connections")
	maxKeys := flag.Int("max-keys", 10, "maximum number of keys before eviction")

	evictionPolicy := flag.String("eviction-policy", "allkeys-lru", "eviction policy")
	evictionRatio := flag.Float64("eviction-ratio", 0.1, "eviction ratio")
	epoolMaxSize := flag.Int("epool-max-size", 16, "eviction pool max size")
	epoolSampleSize := flag.Int("epool-sample-size", 5, "LRU sample size")

	flag.Parse()

	cfg := &Config{
		Protocol:      *protocol,
		Port:          ":" + *port,
		Mode:          ExecutionMode(*mode),
		NumWorkers:    *workers,
		NumIOHandlers: *ioHandlers,
		NumListeners:  *listeners,

		MaxConnection: *maxConn,
		MaxKeyNumber:  *maxKeys,

		EvictionPolicy:     *evictionPolicy,
		EvictionRatio:      *evictionRatio,
		EpoolMaxSize:       *epoolMaxSize,
		EpoolLruSampleSize: *epoolSampleSize,
	}
	if cfg.Mode != SingleThreadMode && cfg.Mode != MultiThreadMode {
		return nil, fmt.Errorf("invalid mode %q, expected single-thread or multi-thread", cfg.Mode)
	}

	if cfg.NumWorkers <= 0 {
		return nil, fmt.Errorf("workers must be greater than 0")
	}

	if cfg.NumIOHandlers <= 0 {
		return nil, fmt.Errorf("io-handlers must be greater than 0")
	}

	if cfg.NumListeners <= 0 {
		return nil, fmt.Errorf("listeners must be greater than 0")
	}

	return cfg, nil
}
