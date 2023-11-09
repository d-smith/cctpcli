package cmd

import (
	"cctpcli/internal/apiclient"
	"cctpcli/internal/eth"
	"cctpcli/internal/types"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/spf13/cobra"
)

// ethBalancesCmd represents the ethBalances command
var ethBalancesCmd = &cobra.Command{
	Use:   "ethBalances [address]",
	Short: "Get the ETH and FIDDY balances of an address",
	Long:  `Get the ETH and FIDDY balances of an address`,
	Args:  cobra.MinimumNArgs(1),
	Run:   getBalancesCmd,
}

func init() {
	rootCmd.AddCommand(ethBalancesCmd)
}

func getBalancesCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("ethBalances requires exactly one argument")
		return
	}
	getBalances(args[0])
}

func balanceInEth(weiBalance *big.Int) *big.Float {
	fbalance := new(big.Float)
	fbalance.SetString(weiBalance.String())
	return new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
}

func getBalances(address string) {
	ethContext := eth.NewEthereumContext()

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

	claims, err := apiclient.GetClaims(address, types.EthereumDomain)
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
