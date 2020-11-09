package filecoin

import (
	"errors"

	filTypes "github.com/filecoin-project/lotus/chain/types"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	sm, ok := signedTx.(*filTypes.SignedMessage)
	if !ok {
		return "", errors.New("tx type assertion error")
	}

	return b.pushMessage(*sm)
}
