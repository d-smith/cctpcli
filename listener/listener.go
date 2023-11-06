package main

import (
	"cctpcli/internal/apiclient"
	"cctpcli/internal/transporter"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	wsURL := os.Getenv("ETH_WS_URL")
	if wsURL == "" {
		log.Fatal("ETH_WS_URL not set")
	}

	client, err := ethclient.Dial(wsURL)
	if err != nil {
		log.Fatal(err)
	}

	deployedAddress := os.Getenv("TRANSPORTER")
	if deployedAddress == "" {
		log.Fatal("TRANSPORTER not set")
	}

	transporterContract, err := transporter.NewTransporter(common.HexToAddress(deployedAddress), client)
	if err != nil {
		log.Fatal(err)
	}
	channel := make(chan *transporter.TransporterMessageSent)
	watchOpts := &bind.WatchOpts{Context: context.Background(), Start: nil}
	sub, err := transporterContract.WatchMessageSent(watchOpts, channel)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening for events...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case ms := <-channel:
			fmt.Println()

			err := apiclient.PostMessageBytesToAttestor(ms.Message, ms.Raw.TxHash.Hex())
			if err != nil {
				log.Println(err)
			}

			fmt.Printf("*** Transaction hash %s ***\n", ms.Raw.TxHash.Hex())

		}
	}
}
