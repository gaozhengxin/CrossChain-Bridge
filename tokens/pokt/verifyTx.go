package pokt

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// VerifyTransaction gets transaction by hash, checks messages, find out swap infos
func (b *Bridge) VerifyTransaction(pairID, txHash string, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	tx, err := b.GetTransaction(txHash)
	if err != nil {
		return nil, errors.Wrap(err, "pokt VerifyTransaction")
	}
	swapInfos, errs := b.verifySwapinTx(tx.(*Tx), true)
	// swapinfos have already aggregated
	for i, swapInfo := range swapInfos {
		if strings.EqualFold(swapInfo.PairID, pairID) {
			return swapInfo, errs[i]
		}
	}
	log.Warn("No such swapInfo")
	return nil, nil
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
func (b *Bridge) verifySwapinTx(tx *Tx, allowUnstable bool) ([]*tokens.TxSwapInfo, []error) {
	// Aggregate swapInfos
	return nil, nil
}
