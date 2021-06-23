package pokt

import (
	"fmt"
	"github.com/pkg/errors"

	tokens "github.com/anyswap/CrossChain-Bridge/tokens"
)

func (b *Bridge) SignTransaction(rawTx interface{}, pairID string) (signedTx interface{}, txHash string, err error) {
	poktRawTx, ok := rawTx.(*Tx)
	if !ok {
		return nil, "", errors.Wrap(ErrPOKTTxType, "pokt SignTransaction")
	}
	fmt.Printf("%v\n", poktRawTx)
	signedTx = new(Tx)
	return signedTx, TxHashString(signedTx.(*Tx)), nil
}

func (b *Bridge) DcrmSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signedTx interface{}, txHash string, err error) {
	poktRawTx, ok := rawTx.(*Tx)
	if !ok {
		return nil, "", errors.Wrap(ErrPOKTTxType, "pokt DcrmSignTransaction")
	}
	fmt.Printf("%v\n", poktRawTx)
	signedTx = new(Tx)
	return signedTx, TxHashString(signedTx.(*Tx)), nil
}
