package types

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
	Signature    string `json:"signature"`
}

const MoonbeamDomain = 2
