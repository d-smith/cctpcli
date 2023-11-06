package main

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	. "cctpcli/internal/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"github.com/holiman/uint256"
	"github.com/mattn/go-sqlite3"
	"github.com/urfave/negroni"
)

var attestorPrivateKey *ecdsa.PrivateKey
var db *sql.DB
var confirmsForAttesation int
var client *ethclient.Client

func init() {
	ethNodeURL := os.Getenv("ETH_URL")
	if ethNodeURL == "" {
		log.Fatal("ETH_NODE_URL not set")
	}

	clientInit, err := ethclient.Dial(ethNodeURL)
	if err != nil {
		log.Fatal(err)
	}

	client = clientInit

	confirmsFromEnv := os.Getenv("ETH_ATTESTOR_CONFIRMS")
	if confirmsFromEnv == "" {
		log.Fatal("ETH_ATTESTOR_CONFIRMS not set")
	}

	confirms, err := strconv.ParseInt(confirmsFromEnv, 10, 32)
	if err != nil {
		log.Fatal(err)
	}

	confirmsForAttesation = int(confirms)

	privateKeyFromEnv := os.Getenv("ETH_ATTESTOR_KEY")
	if privateKeyFromEnv == "" {
		log.Fatal("ETH_ATTESTOR_KEY not set")
	}

	if privateKeyFromEnv[:2] == "0x" {
		privateKeyFromEnv = privateKeyFromEnv[2:]
	}

	attestorPrivateKey, err = crypto.HexToECDSA(privateKeyFromEnv)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing DB...")
	v, _, _ := sqlite3.Version()
	log.Println("Opening sqlite with driver version", v)

	db, err = sql.Open("sqlite3", "attestor.db")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	log.Println("Starting attestor API...")
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/attestor/attest/{txnid}", storeAttestation).Methods("POST")
	r.HandleFunc("/api/v1/attestor/receipts/{destDomain}/{recipient}", listReceipts).Methods("GET")
	r.HandleFunc("/api/v1/attestor/receipts/{destDomain}/{recipient}", listReceipts).Methods("GET")
	r.HandleFunc("/api/v1/attestor/attesations/{id}", getAttestation).Methods("GET")
	r.HandleFunc("/api/v1/attestor/receipts/setspent/{id}", setSpent).Methods("PUT")
	n := negroni.Classic()
	n.UseHandler(r)

	log.Println("Listening on port 3010...")
	err := http.ListenAndServe(":3010", n)
	if err != nil {
		log.Fatalln("Error starting server", err)
	}
}

func isConfirmed(txnHash string) (bool, error) {
	fmt.Printf("Checking if %s is confirmed\n", txnHash)
	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(txnHash))
	if err != nil {
		return false, err
	}

	txnBlockNo := receipt.BlockNumber.Uint64()
	fmt.Printf("Transaction block number: %d\n", txnBlockNo)

	currentBlockNo, err := client.BlockNumber(context.Background())
	if err != nil {
		return false, err
	}

	fmt.Printf("Current block number: %d\n", currentBlockNo)

	return (currentBlockNo - txnBlockNo) >= uint64(confirmsForAttesation), nil
}

func getAttestation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var message string
	var txnid string
	row := db.QueryRow("SELECT message, txnid FROM attestations WHERE id = ?", id)
	err := row.Scan(&message, &txnid)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	confirmed, err := isConfirmed(txnid)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	if !confirmed {
		http.Error(w, "Attestation not confirmed", 404)
		return
	}

	messageBytes, err := hexutil.Decode(string(message))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	signature, err := signMessage(messageBytes)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	sigAsHex := hexutil.Encode(signature)

	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sigAsHex))

}

func storeAttestation(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Called storeAttestation")

	params := mux.Vars(r)
	txnid := params["txnid"]

	fmt.Printf("storing message data for txnid: %s\n", txnid)

	defer r.Body.Close()

	message, err := io.ReadAll(r.Body)
	if err != nil {
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
	}

	parsedMessage, err := parseMessageSent(message)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	printMessage(parsedMessage)

	/*
		signature, err := signMessage(message)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
	*/
	res, err := db.Exec("INSERT INTO attestations (nonce, sender, receiver, source_domain, dest_domain, amount, message, txnid)	 VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		parsedMessage.Nonce, parsedMessage.Sender, parsedMessage.Recipient,
		parsedMessage.LocalDomain, parsedMessage.RemoteDomain,
		parsedMessage.BurnMessage.Amount,
		hexutil.Encode(message), txnid)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Println("Error getting last insert id", err)
	} else {
		log.Println("Inserted row with id", int(id))
	}
}

func signMessage(message []byte) ([]byte, error) {
	msgHash := crypto.Keccak256Hash(message)
	stamp := []byte("\x19Ethereum Signed Message:\n32")
	signature, err := crypto.Sign(crypto.Keccak256Hash(stamp, msgHash.Bytes()).Bytes(), attestorPrivateKey)
	if err != nil {
		return nil, err
	}

	if signature[crypto.RecoveryIDOffset] == 0 || signature[crypto.RecoveryIDOffset] == 1 {
		signature[crypto.RecoveryIDOffset] += 27
	}
	return signature, nil
}

func parseMessageSent(messageSent []byte) (*MessageSent, error) {
	if len(messageSent) < 84 {
		return nil, fmt.Errorf("invalid message sent length")
	}

	messageVersion := binary.BigEndian.Uint32(messageSent[0:4])
	localDomain := binary.BigEndian.Uint32(messageSent[4:8])
	remoteDomain := binary.BigEndian.Uint32(messageSent[8:12])
	nonce := binary.BigEndian.Uint64(messageSent[12:20])

	sender := messageSent[20:52]
	senderHex := hexutil.Encode(sender[12:32])

	recipient := messageSent[52:84]
	recipientHex := hexutil.Encode(recipient[12:32])

	burnMessage := messageSent[84:]

	parsedBurnMessage, err := parseBurnMessage(burnMessage)
	if err != nil {
		return nil, err
	}

	return &MessageSent{
		MessageVersion: messageVersion,
		LocalDomain:    localDomain,
		RemoteDomain:   remoteDomain,
		Nonce:          nonce,
		Sender:         strings.ToLower(senderHex),
		Recipient:      strings.ToLower(recipientHex),
		BurnMessage:    parsedBurnMessage,
	}, nil
}

func parseBurnMessage(burnMessage []byte) (*BurnMessage, error) {
	if len(burnMessage) < 132 {
		return nil, fmt.Errorf("invalid burn message length")
	}

	burnMessageVersion := binary.BigEndian.Uint32(burnMessage[0:4])
	burnToken := burnMessage[4:36]
	mintRecipient := burnMessage[36:68]

	amountBytes := burnMessage[68:100]
	hexAmount := hexutil.Encode(amountBytes[12:32])

	// Convert hexAmount to a uint256

	i := 0
	if hexAmount[0:2] == "0x" {
		i = 2
	}

	for ; i < len(hexAmount); i++ {
		if hexAmount[i] != '0' {
			break
		}
	}

	if i == len(hexAmount) {
		hexAmount = "0x0"
	} else {
		hexAmount = fmt.Sprintf("0x%s", hexAmount[i:])
	}

	amountDec, err := uint256.FromHex(hexAmount)
	if err != nil {
		return nil, err
	}

	sender := burnMessage[100:132]

	return &BurnMessage{
		Version:       burnMessageVersion,
		BurnToken:     hexutil.Encode(burnToken[12:32]),
		MintRecipient: strings.ToLower(hexutil.Encode(mintRecipient[12:32])),
		Amount:        amountDec,
		Sender:        strings.ToLower(hexutil.Encode(sender[12:32])),
	}, nil
}

func printMessage(m *MessageSent) {
	fmt.Printf("Message fields\n")
	fmt.Printf("Message Version: %d\n", m.MessageVersion)
	fmt.Printf("Local domain: %d\n", m.LocalDomain)
	fmt.Printf("Remote domain: %d\n", m.RemoteDomain)
	fmt.Printf("Nonce: %d\n", m.Nonce)
	fmt.Printf("Sender: %s\n", m.Sender)
	fmt.Printf("Recipient: %s\n", m.Recipient)
	fmt.Printf("Burn message:\n")
	fmt.Printf("  Version: %d\n", m.BurnMessage.Version)
	fmt.Printf("  Burn token: %s\n", m.BurnMessage.BurnToken)
	fmt.Printf("  Mint recipient: %s\n", m.BurnMessage.MintRecipient)
	fmt.Printf("  Amount: %d\n", m.BurnMessage.Amount)
	fmt.Printf("  Sender: %s\n", m.BurnMessage.Sender)
}

func setSpent(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id := params["id"]

	if id == "" {
		http.Error(w, "id must be specified", 400)
		return
	}

	idVal, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		http.Error(w, "id must be a number", 400)
		return
	}

	res, err := db.Exec("UPDATE attestations SET spent = true WHERE id = ?", idVal)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "id not found", 404)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func listReceipts(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	destDomain := params["destDomain"]
	recipient := strings.ToLower(params["recipient"])

	if (destDomain == "") || (recipient == "") {
		http.Error(w, "destDomain and recipient must be specified", 400)
		return
	}

	ddVal, err := strconv.ParseUint(destDomain, 10, 32)
	if err != nil {
		http.Error(w, "destDomain must be a number", 400)
		return
	}

	rows, err := db.Query("SELECT id, spent, nonce, sender, receiver, source_domain, dest_domain, amount, message, txnid FROM attestations WHERE dest_domain = ? AND receiver = ?", ddVal, recipient)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	defer rows.Close()

	var receipts []Receipt

	for rows.Next() {
		var receipt Receipt

		err = rows.Scan(&receipt.Id, &receipt.Spent, &receipt.Nonce, &receipt.Sender, &receipt.Recipient, &receipt.SourceDomain, &receipt.DestDomain, &receipt.Amount, &receipt.Message, &receipt.TxnId)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		receipts = append(receipts, receipt)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(receipts)
}
