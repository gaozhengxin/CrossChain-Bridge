package pokt

import (
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

func (b *Bridge) BuildRawTransaction(args *tokens.BuildTxArgs) (rawTx interface{}, err error) {
	rawTx = new(Tx)
	return rawTx, nil
}
