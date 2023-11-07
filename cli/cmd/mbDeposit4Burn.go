/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"cctpcli/internal/eth"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
)

// mbDeposit4BurnCmd represents the mbDeposit4Burn command
var mbDeposit4BurnCmd = &cobra.Command{
	Use:   "mbDeposit4Burn [private key] [recipient] [amount]",
	Short: "Deposit Fiddy to the Transporter contract for burning",
	Long:  `Deposit Fiddy to the Transporter contract for burning. This is the amount of Fiddy that the Transporter contract is allowed to burn on behalf of the address.`,
	Args:  cobra.MinimumNArgs(3),
	Run:   runMBDeposit4BurnCmd,
}

func init() {
	rootCmd.AddCommand(mbDeposit4BurnCmd)
}

func runMBDeposit4BurnCmd(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Println("mbDeposit4Burn requires exactly two arguments")
		return
	}
	mbDeposit4Burn(args[0], args[1], args[2])
}

func mbDeposit4Burn(senderPrivateKey string, recipient string, amount string) {
	mbContext := eth.NewMBEthereumContext()

	approveAmount := new(big.Int)
	approveAmount.SetString(amount, 10)
	txnid, err := mbContext.DepositForBurn(senderPrivateKey, approveAmount, recipient)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Deposited %s: txn hash %s\n", amount, txnid)
}
