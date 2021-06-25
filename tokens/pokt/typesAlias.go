package pokt

import (
	sdk "github.com/pokt-network/pocket-core/types"
	ctypes "github.com/tendermint/tendermint/types"
	//authtypes "github.com/pokt-network/pocket-core/auth/types"
	//tmTypes "github.com/tendermint/tendermint/types"
)

type Block = ctypes.Block

type Tx = sdk.Tx

func StdSignBytes(tx Tx) ([]byte, error) {
	// TODO
	// authtypes.StdSignBytes
	return nil, nil
}

func TxHashString(tx Tx) string {
	// TODO
	/*txBz
	txHash := tmTypes.Tx(txBz).Hash()
	return fmt.Sprintf("%x", txHash)*/
	return ""
}
