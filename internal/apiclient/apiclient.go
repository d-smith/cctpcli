package apiclient

import (
	"bytes"
	. "cctpcli/internal/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var apiHost string
var apiPort int

func init() {
	apiHost = "localhost"
	apiPort = 3010
}

func GetClaims(address string, destinationDomain int) ([]Receipt, error) {

	var receipts []Receipt
	url := fmt.Sprintf("http://%s:%d/api/v1/attestor/receipts/%d/%s",
		apiHost, apiPort, destinationDomain, address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return receipts, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return receipts, err
	}

	if resp.StatusCode != 200 {
		return receipts, fmt.Errorf("error obtaining claims for address")
	}

	err = json.NewDecoder(resp.Body).Decode(&receipts)
	if err != nil {
		return receipts, err
	}

	return receipts, nil

}

func RetrieveAttestation(claimId int) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/attestor/attesations/%d", apiHost, apiPort, claimId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error retrieving attestation")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func SetClaimAsSpent(claimId int) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/attestor/receipts/setspent/%d", apiHost, apiPort, claimId)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("error updating claim status")
	}

	return nil
}

func PostMessageBytesToAttestor(messageBytes []byte, txnHash string) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/attestor/attest/%s", apiHost, apiPort, txnHash)
	fmt.Println("Posting to", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(messageBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error storing attestation")
	}

	return nil
}
