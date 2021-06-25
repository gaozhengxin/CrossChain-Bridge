package pokt

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/anyswap/CrossChain-Bridge/dcrm"
	"github.com/anyswap/CrossChain-Bridge/log"
	tokens "github.com/anyswap/CrossChain-Bridge/tokens"
)

func (b *Bridge) SignTransaction(rawTx interface{}, pairID string) (signedTx interface{}, txHash string, err error) {
	poktRawTx, ok := rawTx.(Tx)
	if !ok {
		return nil, "", errors.Wrap(ErrPOKTTxType, "pokt SignTransaction")
	}

	txBz, err := StdSignBytes(poktRawTx)
	if err != nil {
		return nil, "", err
	}
	privKey := b.GetTokenConfig(pairID).GetDcrmAddressED25519PrivateKey()
	sig := ed25519.Sign(*privKey, txBz)
	fmt.Printf("sig: %x", sig)
	keyID, rsvs, err := dcrm.DoSignOne(b.GetDcrmPublicKey(pairID), fmt.Sprintf("%x", txBz), "")
	// TODO signedTx
	fmt.Printf("KeyID: %v, rsvs: %v\n", keyID, rsvs)

	return signedTx, TxHashString(signedTx.(Tx)), nil
}

func (b *Bridge) DcrmSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signedTx interface{}, txHash string, err error) {
	poktRawTx, ok := rawTx.(Tx)
	if !ok {
		return nil, "", errors.Wrap(ErrPOKTTxType, "pokt DcrmSignTransaction")
	}
	tx, err := b.verifyTransactionWithArgs(poktRawTx, args)
	if err != nil {
		return nil, "", err
	}
	txBz, err := StdSignBytes(tx)
	if err != nil {
		return nil, "", err
	}
	jsondata, _ := json.Marshal(args)
	msgContext := string(jsondata)

	log.Info(b.ChainConfig.BlockChain+" DcrmSignTransaction start", "msghash", fmt.Sprintf("%x", txBz), "txid", args.SwapID)
	keyID, rsvs, err := dcrm.DoSignED25519One(b.GetDcrmPublicKey(args.PairID), fmt.Sprintf("%x", txBz), msgContext)
	if err != nil {
		return nil, "", err
	}
	log.Info(b.ChainConfig.BlockChain+" DcrmSignTransaction finished", "keyID", keyID, "msghash", TxHashString(tx), "txid", args.SwapID)

	if len(rsvs) != 1 {
		return nil, "", fmt.Errorf("get sign status require one rsv but have %v (keyID = %v)", len(rsvs), keyID)
	}

	rsv := rsvs[0]
	log.Trace(b.ChainConfig.BlockChain+" DcrmSignTransaction get rsv success", "keyID", keyID, "txid", args.SwapID, "rsv", rsv)

	// TODO signedTx
	return signedTx, TxHashString(signedTx.(Tx)), nil
}

func (b *Bridge) verifyTransactionWithArgs(tx Tx, args *tokens.BuildTxArgs) (Tx, error) {
	return nil, nil
}
