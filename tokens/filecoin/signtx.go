package filecoin

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anyswap/CrossChain-Bridge/dcrm"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tools/crypto"

	filCrypto "github.com/filecoin-project/go-state-types/crypto"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/minio/blake2b-simd"
)

const (
	retryGetSignStatusCount    = 70
	retryGetSignStatusInterval = 10 * time.Second
)

func (b *Bridge) verifyTransactionWithArgs(tx filTypes.Message, args *tokens.BuildTxArgs) error {
	tokenCfg := b.GetTokenConfig(args.PairID)
	if tokenCfg == nil {
		return fmt.Errorf("[sign] verify tx with unknown pairID '%v'", args.PairID)
	}
	checkReceiver := tokenCfg.ContractAddress
	if args.SwapType == tokens.SwapoutType {
		checkReceiver = args.Bind
	}
	if !strings.EqualFold(tx.To.String(), checkReceiver) {
		return fmt.Errorf("[sign] verify tx receiver failed")
	}
	return nil
}

// DcrmSignTransaction dcrm sign raw tx
func (b *Bridge) DcrmSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signTx interface{}, txHash string, err error) {
	// ==== gzx
	msg, ok := rawTx.(filTypes.Message)
	if !ok {
		return nil, "", errors.New("raw tx type assertion error")
	}

	err = b.verifyTransactionWithArgs(msg, args)
	if err != nil {
		return nil, "", err
	}

	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, "", fmt.Errorf("filecoin message to storage block error: %v", err)
	}

	msgBytes := mb.Cid().Bytes()
	bb := blake2b.Sum256(msgBytes)
	msgHash := hex.EncodeToString(bb[:])
	// ====

	rootPubkey, err := b.prepareDcrmSign(args)
	if err != nil {
		return nil, "", err
	}

	jsondata, _ := json.Marshal(args)
	msgContext := string(jsondata)
	rpcAddr, keyID, err := dcrm.DoSignOne(rootPubkey, args.InputCode, msgHash, msgContext)
	if err != nil {
		return nil, "", err
	}
	log.Info(b.ChainConfig.BlockChain+" DcrmSignTransaction start", "keyID", keyID, "msghash", msgHash, "txid", args.SwapID)
	time.Sleep(retryGetSignStatusInterval)

	var rsv string
	i := 0
	for ; i < retryGetSignStatusCount; i++ {
		signStatus, err2 := dcrm.GetSignStatus(keyID, rpcAddr)
		if err2 == nil {
			if len(signStatus.Rsv) != 1 {
				return nil, "", fmt.Errorf("get sign status require one rsv but have %v (keyID = %v)", len(signStatus.Rsv), keyID)
			}

			rsv = signStatus.Rsv[0]
			break
		}
		switch err2 {
		case dcrm.ErrGetSignStatusFailed, dcrm.ErrGetSignStatusTimeout:
			return nil, "", err2
		}
		log.Warn("retry get sign status as error", "err", err2, "txid", args.SwapID, "keyID", keyID, "bridge", args.Identifier, "swaptype", args.SwapType.String())
		time.Sleep(retryGetSignStatusInterval)
	}
	if i == retryGetSignStatusCount || rsv == "" {
		return nil, "", errors.New("get sign status failed")
	}

	log.Trace(b.ChainConfig.BlockChain+" DcrmSignTransaction get rsv success", "keyID", keyID, "rsv", rsv)

	// signedTx, txHash
	signedTx, err := makeSignedTransaction([]string{rsv}, rawTx)

	txHash = msg.Cid().String()

	log.Info(b.ChainConfig.BlockChain+" DcrmSignTransaction success", "keyID", keyID, "txhash", txHash)
	return signedTx, txHash, err
}

// SignTransaction sign tx with pairID
func (b *Bridge) SignTransaction(rawTx interface{}, pairID string) (signTx interface{}, txHash string, err error) {
	privKey := b.GetTokenConfig(pairID).GetDcrmAddressPrivateKey()
	return b.SignTransactionWithPrivateKey(rawTx, privKey)
}

// SignTransactionWithPrivateKey sign tx with ECDSA private key
func (b *Bridge) SignTransactionWithPrivateKey(rawTx interface{}, privKey *ecdsa.PrivateKey) (signTx interface{}, txHash string, err error) {
	msg, ok := rawTx.(*filTypes.Message)
	if !ok {
		return nil, "", tokens.ErrWrongRawTx
	}
	mb, err := msg.ToStorageBlock()
	if err != nil {
		return nil, "", fmt.Errorf("filecoin message to storage block error: %v", err)
	}

	msgBytes := mb.Cid().Bytes()
	bb := blake2b.Sum256(msgBytes)

	sig, err := crypto.Sign(bb[:], privKey)
	if err != nil {
		return nil, "", fmt.Errorf("sign tx failed, %v", err)
	}
	rsv := hex.EncodeToString(sig)

	signedTx, err := makeSignedTransaction([]string{rsv}, msg)
	if err != nil {
		return nil, "", fmt.Errorf("sign tx failed, %v", err)
	}

	txHash = msg.Cid().String()
	log.Info(b.ChainConfig.BlockChain+" SignTransaction success", "txhash", txHash)
	return signedTx, txHash, err
}

// SignedMessage
func makeSignedTransaction(rsv []string, tx interface{}) (signedTransaction interface{}, err error) {
	sig, err := hex.DecodeString(rsv[0])
	if err != nil {
		return nil, err
	}
	s := filCrypto.Signature{
		Type: filCrypto.SigTypeSecp256k1,
		Data: sig,
	}

	msg, ok := tx.(*filTypes.Message)
	if !ok {
		return nil, fmt.Errorf("tx type assertion error")
	}

	sm := &filTypes.SignedMessage{
		Message:   *msg,
		Signature: s,
	}

	return sm, nil
}

func (b *Bridge) prepareDcrmSign(args *tokens.BuildTxArgs) (rootPubkey string, err error) {
	rootPubkey = b.GetDcrmPublicKey(args.PairID)

	signerAddr := args.From
	if signerAddr == "" {
		token := b.GetTokenConfig(args.PairID)
		signerAddr = token.DcrmAddress
	}

	if args.InputCode != "" {
		childPubkey, err := dcrm.GetBip32ChildKey(rootPubkey, args.InputCode)
		if err != nil {
			return "", err
		}
		signerAddr, err = b.PublicKeyToAddress(childPubkey)
		if err != nil {
			return "", err
		}
	}

	if args.From == "" {
		args.From = signerAddr
	} else if !strings.EqualFold(args.From, signerAddr) {
		log.Error("dcrm sign sender mismath", "inputCode", args.InputCode, "have", args.From, "want", signerAddr)
		return rootPubkey, fmt.Errorf("dcrm sign sender mismath")
	}
	return rootPubkey, nil
}
