package cmd

import (
	"cctpcli/internal/apiclient"
	"cctpcli/internal/eth"
	"cctpcli/internal/types"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// mbBalancesCmd represents the mbBalances command
var mbBalancesCmd = &cobra.Command{
	Use:   "mbBalances [account]",
	Short: "Get the Moonbeam ETH (GLMR?) and FIDDY balances of an address",
	Long:  `Get the Moonbeam ETH (GLMR?) and FIDDY balances of an address`,
	Args:  cobra.MinimumNArgs(1),
	Run:   getMBBalancesCmd,
}

func init() {
	rootCmd.AddCommand(mbBalancesCmd)
}

func getMBBalancesCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("mbBalances requires exactly one argument")
		return
	}
	getMBBalances(args[0])
}

func getMBBalances(address string) {
	ethContext := eth.NewMBEthereumContext()

	bal, err := ethContext.GetBalance(address)
	if err != nil {
		log.Fatal(err)
	}

	eb := balanceInEth(bal)
	fmt.Printf("ETH Balance: %s\n", eb.String())

	fiddyBal, err := ethContext.GetFiddyBalance(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fiddy Balance: %s\n", fiddyBal.String())

	claims, err := apiclient.GetClaims(address, types.MoonbeamDomain)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(claims) == 0 {
		fmt.Println("No claims found for address")
		return
	}

	fmt.Printf("Claims:\n")
	for _, claim := range claims {
		if claim.Spent == false {
			fmt.Printf("  Claim id %d :: Source domain %d -> Destination domain %d, Claimable Amount %d\n", claim.Id, claim.SourceDomain, claim.DestDomain, claim.Amount)
		}
	}

}
