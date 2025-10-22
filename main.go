package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const chainlinkFeed = "0x5f4ec3Df9cbd43714FE2740f5E3616155C5b8419"

var aggregatorABI = `[{
	"inputs": [],
	"name": "latestRoundData",
	"outputs": [
	  { "name": "roundId", "type": "uint80" },
	  { "name": "answer", "type": "int256" },
	  { "name": "startedAt", "type": "uint256" },
	  { "name": "updatedAt", "type": "uint256" },
	  { "name": "answeredInRound", "type": "uint80" }
	],
	"stateMutability": "view",
	"type": "function"
}]`

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <eth-address>")
	}

	addr := common.HexToAddress(os.Args[1])

	client, err := ethclient.Dial("https://eth-mainnet.g.alchemy.com/v2/JJL-xKNzxgcmvEPuMrbDB")
	if err != nil {
		log.Fatalf("RPC connect error: %v", err)
	}
	defer client.Close()

	balanceWei, err := client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		log.Fatalf("Cannot get balance: %v", err)
	}

	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balanceWei), big.NewFloat(1e18))

	parsedABI, err := abi.JSON(strings.NewReader(aggregatorABI))
	if err != nil {
		log.Fatalf("ABI parse error: %v", err)
	}

	feedAddr := common.HexToAddress(chainlinkFeed)
	contract := bind.NewBoundContract(feedAddr, parsedABI, client, client, client)

	var out []interface{}
	err = contract.Call(nil, &out, "latestRoundData")
	if err != nil {
		log.Fatalf("Chainlink call error: %v", err)
	}

	price := out[1].(*big.Int)
	priceFloat := new(big.Float).Quo(new(big.Float).SetInt(price), big.NewFloat(1e8))

	usdValue := new(big.Float).Mul(balanceEth, priceFloat)

	fmt.Printf("Address: %s\n", addr.Hex())
	fmt.Printf("ETH balance: %f\n", balanceEth)
	fmt.Printf("ETH/USD: %f\n", priceFloat)
	fmt.Printf("USD equivalent: $%.2f\n", usdValue)
}
