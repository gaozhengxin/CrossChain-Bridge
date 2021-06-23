package pokt

import (
	tokens "github.com/anyswap/CrossChain-Bridge/tokens"
)

func (b *Bridge) SignTransaction(rawTx interface{}, pairID string) (signedTx interface{}, txHash string, err error) {
	return nil, "", nil
}

func (b *Bridge) DcrmSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signedTx interface{}, txHash string, err error) {
	return nil, "", nil
}
