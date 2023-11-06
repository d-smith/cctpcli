package types

import "github.com/holiman/uint256"

type Receipt struct {
	Id           int    `json:"id"`
	Spent        bool   `json:"spent"`
	Nonce        uint64 `json:"nonce"`
	Sender       string `json:"sender"`
	Recipient    string `json:"recipient"`
	SourceDomain uint32 `json:"sourceDomain"`
	DestDomain   uint32 `json:"destDomain"`
	Amount       uint64 `json:"amount"`
	Message      string `json:"message"`
	TxnId        string `json:"txnid"`
}

const MoonbeamDomain = 2

type MessageSent struct {
	MessageVersion uint32
	LocalDomain    uint32
	RemoteDomain   uint32
	Nonce          uint64
	Sender         string
	Recipient      string
	BurnMessage    *BurnMessage
}

type BurnMessage struct {
	Version       uint32
	BurnToken     string
	MintRecipient string
	Amount        *uint256.Int
	Sender        string
}
