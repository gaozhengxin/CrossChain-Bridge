package filecoin

import (
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"
)

func (b *Bridge) processTransaction(txid string) {
	if b.IsSrc {
		b.processSwapin(txid)
	}
}

func (b *Bridge) processSwapin(txid string) {
	swapInfo, err := b.verifySwapinTx(PairID, txid, true)
	if swapInfo == nil {
		return
	}
	tools.RegisterSwapin(txid, []*tokens.TxSwapInfo{swapInfo}, []error{err})
}
