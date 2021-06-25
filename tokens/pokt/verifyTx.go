package pokt

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// VerifyTransaction gets transaction by hash, checks messages, find out swap infos
func (b *Bridge) VerifyTransaction(pairID, txHash string, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	tx, err := b.GetTransaction(txHash)
	if err != nil {
		return nil, errors.Wrap(err, "pokt VerifyTransaction")
	}
	return b.verifySwapinTx(tx.(Tx), true)
}

func (b *Bridge) VerifyMsgHash(rawTx interface{}, msgHash []string) error {
	poktTx, ok := rawTx.(*Tx)
	if !ok {
		return errors.Wrap(ErrPOKTTxType, "pokt VerifyMsgHash")
	}
	fmt.Printf("%v\n", poktTx)
	// Calculate tx signature payload
	return nil
}

// verifySwapinTx
// TODO define tx struct type
func (b *Bridge) verifySwapinTx(tx Tx, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	// Check every message
	// 1. message is excuted successfully
	// 2. transfer to our deposit address
	// 3. get bind address
	// 4. make a swapinfo
	swapinfo := &tokens.TxSwapInfo{
		PairID:    pairID,
		Hash:      TxHashString(tx),
		Height:    0,
		Timestamp: 0,
		From:      "",
		TxTo:      "",
		To:        "",
		Bind:      "",
		Value:     big.NewInt(0),
	}
	return swapinfo, nil
}
