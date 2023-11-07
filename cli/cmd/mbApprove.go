/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"cctpcli/internal/eth"
	"math/big"

	"github.com/spf13/cobra"
)

// mbApproveCmd represents the mbApprove command
var mbApproveCmd = &cobra.Command{
	Use:   "mbApprove [sender private key] [amount]",
	Short: "Approve a token allowance for the Transporter",
	Long:  `Approve a token allowance for the Transporter. This is the amount of Fiddy that the Transporter contract is allowed to burn on behalf of the address.`,
	Args:  cobra.MinimumNArgs(2),
	Run:   mbFiddyApproveCmd,
}

func init() {
	rootCmd.AddCommand(mbApproveCmd)
}

func mbFiddyApproveCmd(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println("mbApprove requires exactly two arguments")
		return
	}
	mbFiddyApprove(args[0], args[1])
}

func mbFiddyApprove(senderPrivateKey string, amount string) {
	mbContext := eth.NewMBEthereumContext()

	approveAmount := new(big.Int)
	approveAmount.SetString(amount, 10)
	txnid, err := mbContext.Approve(senderPrivateKey, approveAmount)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Approved %s: txn id %s\n", amount, txnid)
}
