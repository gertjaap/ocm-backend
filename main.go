package main

import (
	"database/sql"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gertjaap/ocm-backend/http"
	"github.com/gertjaap/ocm-backend/logging"
	"github.com/gertjaap/ocm-backend/processor"
	_ "github.com/lib/pq"
)

func main() {
	logging.SetLogLevel(int(logging.LogLevelInfo))
	rpc, err := initRPC()
	if err != nil {
		panic(err)
	}

	connStr := os.Getenv("PGSQL_CONNECTION")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	p, err := processor.NewProcessor(rpc, db)
	if err != nil {
		panic(err)
	}

	h := http.NewHttpServer(db, p)

	go p.ProcessLoop()
	h.Run()
}

func initRPC() (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv("RPCHOST"),
		User:         os.Getenv("RPCUSER"),
		Pass:         os.Getenv("RPCPASS"),
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	logging.Debugf("RPC Server: %s", connCfg.Host)
	return rpcclient.New(connCfg, nil)
}
