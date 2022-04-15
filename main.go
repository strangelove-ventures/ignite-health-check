package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	cosmosClient "github.com/cosmos/cosmos-sdk/client"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

const RPCTimeoutSeconds = 5

func newClient(addr string) (rpcclient.Client, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}

	httpClient.Timeout = 10 * time.Second
	rpcClient, err := rpchttp.NewWithClient(addr, "", httpClient)
	if err != nil {
		return nil, err
	}

	return rpcClient, nil
}

func getCosmosClient(rpcAddress string) (*cosmosClient.Context, error) {
	client, err := newClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	return &cosmosClient.Context{
		Client: client,
		Input:  os.Stdin,
		Output: os.Stdout,
	}, nil
}

func inSync(client *cosmosClient.Context) bool {
	node, err := client.GetNode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting node: %v\n", err)
		return false
	}
	statusCtx, statusCtxCancel := context.WithTimeout(context.Background(), time.Duration(time.Second*RPCTimeoutSeconds))
	defer statusCtxCancel()
	status, err := node.Status(statusCtx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		return false
	}

	return !status.SyncInfo.CatchingUp
}

type InSyncResponse struct {
	Address string `json:"address"`
	InSync  bool   `json:"in_sync"`
}

func main() {
	rpcAddress := os.Getenv("RPC_ADDRESS")
	if rpcAddress == "" {
		rpcAddress = "tcp://localhost:26657"
	}
	client, err := getCosmosClient(rpcAddress)
	if err != nil {
		panic(err)
	}

	inSyncResponse, err := json.Marshal(InSyncResponse{Address: rpcAddress, InSync: true})
	if err != nil {
		panic(err)
	}
	notInSyncResponse, err := json.Marshal(InSyncResponse{Address: rpcAddress, InSync: false})
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var status int
		var response []byte
		if inSync(client) {
			status = http.StatusOK
			response = inSyncResponse
		} else {
			status = http.StatusServiceUnavailable
			response = notInSyncResponse
		}
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "1251"
	}
	listenAddr := fmt.Sprintf(":%s", port)

	fmt.Printf("Health check for %s listening on port %s\n", rpcAddress, port)

	panic(http.ListenAndServe(listenAddr, nil))
}
