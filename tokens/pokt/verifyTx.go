package pokt

import (
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

func (b *Bridge) VerifyTransaction(pairID, txHash string, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	return nil, nil
}

func (b *Bridge) VerifyMsgHash(rawTx interface{}, msgHash []string) error {
	return nil
}

// verifySwapinTx
// TODO define tx struct type
func (b *Bridge) verifySwapinTx(tx interface{}, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	return nil, nil
}