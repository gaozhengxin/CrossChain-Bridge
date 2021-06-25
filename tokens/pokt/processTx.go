package pokt

import (
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"
)

// processTransaction
func (b *Bridge) processTransaction(tx interface{}) {
	poktTx, ok := tx.(Tx)
	if !ok {
		log.Warn("process pokt transaction error, transaction type error")
		return
	}
	swapInfo, err := b.verifySwapinTx(poktTx, true)
	tools.RegisterSwapin(TxHashString(poktTx), []*tokens.TxSwapInfo{swapInfo}, []error{err})
	return
}
