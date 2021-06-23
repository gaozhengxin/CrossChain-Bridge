package pokt

import (
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"
)

// processTransaction
func (b *Bridge) processTransaction(tx interface{}) {
	poktTx, ok := tx.(*Tx)
	if !ok {
		log.Warn("process pokt transaction error, transaction type error")
		return
	}
	swapInfos, errs := b.verifySwapinTx(poktTx, true)
	if swapInfos != nil && len(swapInfos) > 0 {
		tools.RegisterSwapin(TxHashString(poktTx), swapInfos, errs)
	}
	return
}
