package eos

import (
	"context"
	"errors"
	"fmt"

	"github.com/anyswap/CrossChain-Bridge/log"
	eosgo "github.com/eoscanada/eos-go"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	cli := b.GetClient()
	ctx := context.Background()

	stx, ok := signedTx.(*eosgo.SignedTransaction)
	if !ok {
		return "", errors.New("tx type assertion error")
	}

	packedTx, err := stx.Pack(eosgo.CompressionNone)
	if err != nil {
		return "", err
	}
	log.Info("SendTransaction", "packedTx", fmt.Sprintf("%+v", packedTx))

	hashBytes, err := packedTx.ID()
	if err != nil {
		return "", err
	}
	txHash = hashBytes.String()

	res, err := cli.PushTransaction(ctx, packedTx)
	if err != nil {
		return txHash, err
	}
	if res.TransactionID != txHash {
		return txHash, fmt.Errorf("push transaction returns wrong txid")
	}
	return txHash, nil
}
