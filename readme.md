# cctpcli

This is a demo project that offers a simple set of CLI commands to
put a simplified CCTP through its paces. There are three main components to the project:

1. cli - Simple command line interface to obtain a token and used the CCTP
to deposit/burn and to mint from burned.
2. api - The attestor API that can store burn context and provide services
 to list claims, obtain attestations, and mark items as spent.
 3. listener - A program to listen for depositForBurn events (MessageSent) that can parse the burn context from the event and call an API to store it for retrieving an attestation once the burn is confirmed.

## Prerequisites

### Test Networks

Before running tne CLI, you will need access to Ethereum and Moonbeam dev environments.

* You can use anvil to run an Ethereum local development network.
* You can run a Moonbeam local development network using docker.

The readme for the [CCTP Lite 2](https://github.com/d-smith/cctp-lite-2) project has details on how to run these environments on a developer's machine.

### Smart Contract Deployment

To use the CLI, you will need to deploy the CCTP Lite 2 smart contracts to the Ethereum and Moonbeam test environments. This includes both the FiddyCent ERC20 token and the Transporter contract.

The details for how to do this are provided in the [CCTP Lite 2](https://github.com/d-smith/cctp-lite-2) project.

Note that the deployment address for the FiddyCent token and the Transporter contract on both networks will be needed to run the CLI. The smart contract project includes scripts to extract these addresses from the deployment logs.

### Running the API Server

Prior to running the API server, the database must be created.

```
sqlite3 attestor.db < att.sql
```

With the database created, the following environment variables must be set:

* ETH_URL - The RPC endpoint URL for the Ethereum node
* MOONBEAM_URL - The RPC endpoint URL for the Moonbeam node
* ETH_ATTESTOR_CONFIRMS - the number of Ethereum block confirmations to wait for before considering a burn confirmed. This can be set to zero for convenience when running a local demo.
* ETH_ATTESTOR_PRIVATE_KEY - The private key for the Ethereum account that will be used to sign attestations. You can use a key from the .env file in the cli directory, noting this key must align with the attestor account configure in the Transporter contract on the moonbeam side.
* MB_ATTESTOR_PRIVATE_KEY - The private key for the Moonbeam account that will be used to sign attestations. You can use a key from the .env file in the cli directory, noting this key must align with the attestor account configure in the Transporter contract on the ethereum side.

### Running the Listeners

We deploy two listeners, one for each network. The listeners are configured using environment variables.

Ethereum listener:

* ETH_WS_URL - The websocket URL for the Ethereum node. For example when running a local Ethereum node using anvil, this would be ws://localhost:8544
* TRANSPORTER - The address of the Transporter contract on the Ethereum side. 

Moonbeam listener:

* MB_WS_URL - The websocket URL for the Moonbeam node. For example when running a local Moonbeam node using docker, this would be ws://localhost:9944
* MB_TRANSPORTER - The address of the Transporter contract on the Moonbeam side.





 ## Example of the CLI in Action

 This project assumes the deployment of the [CCTP Lite 2](https://github.com/d-smith/cctp-lite-2) project to Ethereum and Moonbeam test environments - see the readme for that project for details on how to deploy the contracts.


### Transport FiddyCent from Ethereum to Moonbeam

Environment set up for the cli:

* Source the .env file in the cli directory to set up the environment variables for the cli. Note this assumes the use of the prefunded accounts in the Moonbeam and Ethereum dev environments.
* FIDDY_ETH_ADDRESS - FiddyCent contract address on Ethereum
* TRANSPORTER - Transporter contract address on Ethereum
* FIDDY_MB_ADDRESS - FiddyCent contract address on Moonbeam
* MB_TRANSPORTER - Transporter contract address on Moonbeam

Install the cli command: `go install`

This will install the cli binary, the location of which, from the go install documentation:

```
Executables are installed in the directory named by the GOBIN environment
variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH
environment variable is not set. Executables in $GOROOT
are installed in $GOROOT/bin or $GOTOOLDIR instead of $GOBIN.```
```



```console
$ # Show starting balances
$ cli ethBalances $ACCT3
ETH Balance: 10000000000000000000000
Fiddy Balance: 0
No claims found for address

$ cli mbBalances $MBACCT3
ETH Balance: 1208925819614629174706176
Fiddy Balance: 0
No claims found for address
```

```console
$ # Use the faucet to obtain some Fiddy on the Ethereum side
$ cli ethDrip $ACCT3 50
Dripped 50 to 0x7bA7d161F9E8B707694f434d65c218a1F0853B1C: txn id 0x6ceaa362fac94e32e834745d06e50ed9569c5621b9b387603d93a4d8cd85eeed

$ cli ethBalances $ACCT3
ETH Balance: 10000000000000000000000
Fiddy Balance: 50
```

```console
$ # Authorize the transport contract to spend Fiddy on behalf of ACCT1
$ cli ethAllowance $ACCT3
Allowance: 0

$ cli ethApprove $ACCT3KEY 25
Approved 25: txn id 0x8fb8840e347838277d373a76c63a496a840252cfa088af9ce76761e3a5f7ff71

$ cli ethAllowance $ACCT3
Allowance: 25
```

```console
$ # Deposit 5 Fiddy on the Ethereum side for transport to Moonbeam
$ cli ethDeposit4Burn $ACCT3KEY $MBACCT3 5
Deposited 5: txn hash 0x957bfe8a77ef16cd3f67077bc645f24966bffdfdb233d90bb1ebb132372e9325

$ cli ethDeposit4Burn $ACCT3KEY $MBACCT3 10
Deposited 10: txn hash 0x5efcc307b5df7b3312086c538ed14b4d65b02c20df93d933ca46d9b47234ae50
```

```console
$ # View available claims on the Moonbeam side
$ cli mbBalances $MBACCT3
ETH Balance: 1208925819614629174706176
Fiddy Balance: 0
Claims:
  Claim id 1 :: Source domain 1 -> Destination domain 2, Claimable Amount 5
  Claim id 2 :: Source domain 1 -> Destination domain 2, Claimable Amount 10
```

```console
$ # Claim 10 Fiddy on the Moonbeam side
$ cli mbMintFromBurned $MBACCT3 $MBACCT3KEY 2
Minted: txn hash 0x036f2d3cd8500cf4d48c37a89c38cc8556cc466475c9577f6cb50afb655c0f44

$ cli mbBalances $MBACCT3
ETH Balance: 1208925818413956864309442
Fiddy Balance: 10
Claims:
  Claim id 1 :: Source domain 1 -> Destination domain 2, Claimable Amount 5
```

```console
$ # Show final balances
$ cli ethBalances $ACCT3
ETH Balance: 9999999829348848426872
Fiddy Balance: 35
No claims found for address


$ cli mbBalances $MBACCT3
ETH Balance: 1208925818413956864309442
Fiddy Balance: 10
Claims:
  Claim id 1 :: Source domain 1 -> Destination domain 2, Claimable Amount 5
```


### Transport FiddyCent from Moonbeam to Ethereum

```
$ cli mbBalances $MBACCT5
ETH Balance: 1208925819152591000929336
Fiddy Balance: 0
No claims found for address

$ cli ethBalances $ACCT5
ETH Balance: 10000000000000000000000
Fiddy Balance: 0
No claims found for address

$ cli mbDrip $MBACCT5 50
Dripped 50 to 0xC0F0f4ab324C46e55D02D0033343B4Be8A55532d: txn id 0x02aac0b35b8336b8a15c8d1d6ec9889e03b3fd97e4d4e344af56dc1add1ee1d0

$ cli mbBalances $MBACCT5
ETH Balance: 1208925819152591000929336
Fiddy Balance: 50
No claims found for address

$ cli mbApprove $MBACCT5KEY 100
Approved 100: txn id 0x5c2a486bc6ffaa29c3c787aca1cc65fe1a95436c15e2cf26924c3dc2ace88702

$ cli mbAllowance $MBACCT5
Allowance: 100

$ cli mbDeposit4Burn $MBACCT5KEY $ACCT5 5
Deposited 5: txn hash 0xb8503993a157e006768f50b36cd9f2dd8314e2acfc4c0e59e04093434ab6ef38

$ cli ethBalances $ACCT5
ETH Balance: 10000000000000000000000
Fiddy Balance: 0
Claims:
  Claim id 3 :: Source domain 2 -> Destination domain 1, Claimable Amount 5

$ cli ethMintFromBurned $ACCT5 $ACCT5KEY 3
Signature:  0x972e0bee4e6ba75fc7cc48555b203da457cf769bb2e44ce55df8a86c322f5ea36cbcc56328981105cd6d2bac19e5213e85573a5cf3991911d21e34112f52c50b1c
Minted: txn hash 0x56222fbf0248417a46fff017a37e9bda86726e19d2d5d98ecc97890e82f5b9ab
Claim marked as spent
```