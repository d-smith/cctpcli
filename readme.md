# cctpcli

This is a demo project that offers a simple set of CLI commands to
put a simplified CCTP through its paces. There are three main components to the project:

1. cli - Simple command line interface to obtain a token and used the CCTP
to deposit/burn and to mint from burned.
2. api - The attestor API that can store burn context and provide services
 to list claims, obtain attestations, and mark items as spent.
 3. listener - A program to listen for depositForBurn events (MessageSent) that can parse the burn context from the event and call an API to store it for retrieving an attestation once the burn is confirmed.

 ## Example of the CLI in Action

 This project assumes the deployment of the [CCTP Lite 2](https://github.com/d-smith/cctp-lite-2) project to Ethereum and Moonbeam test environments - see the readme for that project for details on how to deploy the contracts.


### Running the event listener

In a dedicated shell, set up the environment and run the ethereum network smart contract event listener:

```console
$ export FIDDY_ETH_ADDRESS=0xC0a4b9e04fB55B1b498c634FAEeb7C8dD5895b53
$ export TRANSPORTER=0xa7F08a6F40a00f4ba0eE5700F730421C5810f848
$ export FIDDY_MB_ADDRESS=0x970951a12F975E6762482ACA81E57D5A2A4e73F4
$ export MB_TRANSPORTER=0x3ed62137c5DB927cb137c26455969116BF0c23Cb
$ . .env
$ go run main.go runEthEventListener
```

The listener subscribes to the MessageSent event emitted by the smart contract when the Transporter burns
a deposit for mint on a remote chain, in this instance Fiddy tokens deposited and burned on the ethereum
side for transport to moonbeam.


### Transport FiddyCent from Ethereum to Moonbeam


```console
$ export FIDDY_ETH_ADDRESS=0xC0a4b9e04fB55B1b498c634FAEeb7C8dD5895b53
$ export TRANSPORTER=0xa7F08a6F40a00f4ba0eE5700F730421C5810f848
$ export FIDDY_MB_ADDRESS=0x970951a12F975E6762482ACA81E57D5A2A4e73F4
$ export MB_TRANSPORTER=0x3ed62137c5DB927cb137c26455969116BF0c23Cb
$ . .env

```

```console
$ # Show starting balances
$ cctpcli ethBalances $ACCT3
ETH Balance: 10000000000000000000000
Fiddy Balance: 0
No claims found for address

$ cctpcli mbBalances $MBACCT3
ETH Balance: 1208925819614629174706176
Fiddy Balance: 0
No claims found for address
```

```console
$ # Use the faucet to obtain some Fiddy on the Ethereum side
$ cctpcli ethDrip $ACCT3 50
Dripped 50 to 0x7bA7d161F9E8B707694f434d65c218a1F0853B1C: txn id 0x6ceaa362fac94e32e834745d06e50ed9569c5621b9b387603d93a4d8cd85eeed

$ cctpcli ethBalances $ACCT3
ETH Balance: 10000000000000000000000
Fiddy Balance: 50
```

```console
$ # Authorize the transport contract to spend Fiddy on behalf of ACCT1
$ cctpcli ethAllowance $ACCT3
Allowance: 0

$ go run main.go ethApprove $ACCT1KEY 25
Approved 25: txn id 0x8fb8840e347838277d373a76c63a496a840252cfa088af9ce76761e3a5f7ff71

$ cctpcli ethAllowance $ACCT3
Allowance: 25
```

```console
$ # Deposit 5 Fiddy on the Ethereum side for transport to Moonbeam
$ cctpcli ethDeposit4Burn $ACCT3KEY $MBACCT3 5
Deposited 5: txn hash 0x957bfe8a77ef16cd3f67077bc645f24966bffdfdb233d90bb1ebb132372e9325

$ cctpcli ethDeposit4Burn $ACCT3KEY $MBACCT3 10
Deposited 10: txn hash 0x5efcc307b5df7b3312086c538ed14b4d65b02c20df93d933ca46d9b47234ae50
```

```console
$ # View available claims on the Moonbeam side
$ cctpcli mbBalances $MBACCT3
ETH Balance: 1208925819614629174706176
Fiddy Balance: 0
Claims:
  Claim id 1 :: Source domain 1 -> Destination domain 2, Claimable Amount 5
  Claim id 2 :: Source domain 1 -> Destination domain 2, Claimable Amount 10
```

```console
$ # Claim 10 Fiddy on the Moonbeam side
$ cctpcli mbMintFromBurned $MBACCT3 $MBACCT3KEY 2
Minted: txn hash 0x036f2d3cd8500cf4d48c37a89c38cc8556cc466475c9577f6cb50afb655c0f44

$ cctpcli mbBalances $MBACCT3
ETH Balance: 1208925818413956864309442
Fiddy Balance: 10
Claims:
  Claim id 1 :: Source domain 1 -> Destination domain 2, Claimable Amount 5
```

```console
$ # Show final balances
$ cctpcli ethBalances $ACCT3
ETH Balance: 9999999829348848426872
Fiddy Balance: 35
No claims found for address


$ cctpcli mbBalances $MBACCT3
ETH Balance: 1208925818413956864309442
Fiddy Balance: 10
Claims:
  Claim id 1 :: Source domain 1 -> Destination domain 2, Claimable Amount 5
```


