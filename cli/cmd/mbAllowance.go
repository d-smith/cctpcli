/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"cctpcli/internal/eth"

	"github.com/spf13/cobra"
)

// mbAllowanceCmd represents the mbAllowance command
var mbAllowanceCmd = &cobra.Command{
	Use:   "mbAllowance [account]",
	Short: "Obtain the Transporter Fiddy allowance for an address",
	Long:  `Obtain the Transporter Fiddy allowance for an address`,
	Args:  cobra.MinimumNArgs(1),
	Run:   getMBAllowanceCmd,
}

func init() {
	rootCmd.AddCommand(mbAllowanceCmd)
}

func getMBAllowanceCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("mbAllowance requires exactly one argument")
		return
	}
	getAllowance(args[0])
}

func getMBAllowance(address string) {
	mbContext := eth.NewMBEthereumContext()

	allowance, err := mbContext.GetAllowance(address)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Allowance: %s\n", allowance.String())
}
