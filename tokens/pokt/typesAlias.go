package pokt

import (
	"fmt"

	"github.com/tendermint/tendermint/types"
)

type Block = types.Block

type Tx = types.Tx

func TxHashString(tx *Tx) string {
	return fmt.Sprintf("%x", tx.Hash())
}
