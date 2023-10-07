package utils

import (
	"database/sql"
	"encoding/base64"
	"flag"
	"os"

	"github.com/sgladkov/tortuga/internal/blockchain"
)

type Config struct {
	Endpoint  string
	LogLevel  string
	DB        *sql.DB
	RpcNode   string
	WalletKey []byte
}

func (sc *Config) Read() error {
	var DatabaseDSN string
	var key string
	flag.StringVar(&sc.Endpoint, "a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	flag.StringVar(&sc.LogLevel, "l", "info", "log level (fatal,  error,  warn, info, debug)")
	flag.StringVar(&sc.RpcNode, "r", "/tmp/metrics-db.json", "file to store ans restore metrics")
	flag.StringVar(&DatabaseDSN, "d", "", "database connection string for PostgreSQL")
	flag.StringVar(&key, "k", "", "key to verify data integrity")
	flag.Parse()

	// check environment
	address := os.Getenv("ADDRESS")
	if len(address) > 0 {
		sc.Endpoint = address
	}
	envLogLevel := os.Getenv("LOG_LEVEL")
	if len(envLogLevel) > 0 {
		sc.LogLevel = envLogLevel
	}
	envDatabaseDSN := os.Getenv("DATABASE_DSN")
	if len(envDatabaseDSN) > 0 {
		DatabaseDSN = envDatabaseDSN
	}
	envRpcNode := os.Getenv("RPC_NODE")
	if len(envRpcNode) > 0 {
		sc.RpcNode = envRpcNode
	}
	envKey := os.Getenv("KEY")
	if len(envKey) > 0 {
		key = envKey
	}

	var err error
	if len(key) > 0 {
		sc.WalletKey, err = base64.StdEncoding.DecodeString(key)
	} else {
		sc.WalletKey, err = blockchain.GeneratePrivateKey()
	}

	sc.DB, err = sql.Open("postgres", DatabaseDSN)

	return err
}
