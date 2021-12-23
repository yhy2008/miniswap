package test

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

func Connect() *ethclient.Client {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		panic(err)
	}

	return client
}
