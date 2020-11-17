package eos

import (
	"errors"
	"fmt"

	eosgo "github.com/eoscanada/eos-go"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	cli := GetClient()

	stx, ok := signedTx.(*eosgo.SignedTransaction)
	if !ok {
		return "", errors.New("tx type assertion error")
	}

	packedTx, err := stx.Pack()
	if err != nil {
		return "", err
	}

	hashBytes, err := packedTx.ID()
	if err != nil {
		return "", err
	}
	txHash = hashBytes.String()

	res, err := cli.PushTransaction(packedTx)
	if err != nil {
		return txHash, err
	}
	if res.TransactionID != txHash {
		return txHash, fmt.Errorf("push transaction returns wrong txid")
	}
	return txHash, nil
}
