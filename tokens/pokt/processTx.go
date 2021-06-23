package pokt

import (
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"
)

// processTransaction
// TODO define tx type
func (b *Bridge) processTransaction(tx interface{}) {
	// TODO txid := tx.Id()
	txid := ""
	swapInfo, err := b.verifySwapinTx(tx, true)
	tools.RegisterSwapin(txid, []*tokens.TxSwapInfo{swapInfo}, []error{err})
	return
}
