/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"cctpcli/internal/apiclient"
	"cctpcli/internal/eth"
	"cctpcli/internal/types"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ethMintFromBurnedCmd represents the ethMintFromBurned command
var ethMintFromBurnedCmd = &cobra.Command{
	Use:   "ethMintFromBurned [receiver address] [receiver key] [claim id]",
	Short: "Mint Fiddy on Ethereum from Fiddy burnt on Moonbeam",
	Long:  `Mint Fiddy on Ethereum from Fiddy burnt on Moonbeam`,
	Args:  cobra.MinimumNArgs(3),
	Run:   runEthMintFromBurnedCmd,
}

func init() {
	rootCmd.AddCommand(ethMintFromBurnedCmd)
}

func runEthMintFromBurnedCmd(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Println("ethMintFromBurned requires exactly 3 arguments")
		return
	}
	ethMintFromBurned(args[0], args[1], args[2])
}

func ethMintFromBurned(receiverAddress, receiverKey, claimId string) {
	ethContext := eth.NewEthereumContext()
	claims, err := apiclient.GetClaims(receiverAddress, types.EthereumDomain)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(claims) == 0 {
		fmt.Println("No claims found for address")
		return
	}

	claimIdInt, err := strconv.ParseInt(claimId, 10, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	var idx = -1
	for pos, claim := range claims {
		if claim.Id == int(claimIdInt) {
			idx = pos
			break
		}
	}

	if idx == -1 {
		fmt.Printf("Claim %d not found for address %s\n", claimIdInt, receiverAddress)
		return
	}

	claim := claims[idx]

	signature, err := apiclient.RetrieveAttestation(claim.Id)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Signature: ", signature)

	txnid, err := ethContext.MintFromBurned(receiverKey, claim.Message, signature)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Minted: txn hash %s\n", txnid)

	err = apiclient.SetClaimAsSpent(claim.Id)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Claim marked as spent")
}
