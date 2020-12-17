package eos

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/anyswap/CrossChain-Bridge/dcrm"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/btcsuite/btcd/btcec"
	"github.com/eoscanada/eos-go/ecc"
)

const (
	retryGetSignStatusCount    = 70
	retryGetSignStatusInterval = 10 * time.Second

	canonicalRetryTimes    = 25
	canonicalRetryInterval = 3 * time.Second
)

func (b *Bridge) verifyTransactionWithArgs(tx *eosgo.Transaction, args *tokens.BuildTxArgs) error {
	/* 	if tx.To() == nil || *tx.To() == (common.Address{}) {
	   		return fmt.Errorf("[sign] verify tx receiver failed")
	   	}
	   	tokenCfg := b.GetTokenConfig(args.PairID)
	   	if tokenCfg == nil {
	   		return fmt.Errorf("[sign] verify tx with unknown pairID '%v'", args.PairID)
	   	}
	   	checkReceiver := tokenCfg.ContractAddress
	   	if args.SwapType == tokens.SwapoutType {
	   		checkReceiver = args.Bind
	   	}
	   	if !strings.EqualFold(tx.To().String(), checkReceiver) {
	   		return fmt.Errorf("[sign] verify tx receiver failed")
	   	} */
	return nil
}

// DcrmSignTransaction dcrm sign raw tx
func (b *Bridge) DcrmSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signTx interface{}, txHash string, err error) {
	eostx, ok := rawTx.(*eosgo.Transaction)
	if !ok {
		return nil, "", errors.New("raw tx type assertion error")
	}
	err = b.verifyTransactionWithArgs(eostx, args)
	if err != nil {
		return nil, "", err
	}

	stx := eosgo.NewSignedTransaction(eostx)

	txdata, cfd, err := stx.PackedTransactionAndCFD()
	if err != nil {
		return nil, "", err
	}
	digest := eosgo.SigDigest(opts.ChainID, txdata, cfd)
	msgHash := hex.EncodeToString(digest)

	/*rootPubkey, err := b.prepareDcrmSign(args)
	if err != nil {
		return nil, "", err
	}*/

	jsondata, _ := json.Marshal(args)
	msgContext := string(jsondata)

	var rsv, keyID string

	for retrytime := 0; retrytime < canonicalRetryTimes; retrytime++ {
		// ======================= Call dcrm sign, check canonical
		//rpcAddr, keyID, err := dcrm.DoSignOne(rootPubkey, args.InputCode, msgHash, msgContext)
		rpcAddr, keyID, err := dcrm.DoSignOne(b.GetDcrmPublicKey(args.PairID), msgHash, msgContext)
		if err != nil {
			return nil, "", err
		}
		log.Info(b.ChainConfig.BlockChain+" DcrmSignTransaction start", "keyID", keyID, "msghash", msgHash, "txid", args.SwapID)
		time.Sleep(retryGetSignStatusInterval)

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

		sig, decodeErr := hex.DecodeString(rsv)
		if decodeErr != nil {
			return nil, "", errors.New("invalid rsv")
		}
		if IsCanonical(sig) == true {
			break
		}
		rsv = ""
		time.Sleep(canonicalRetryInterval)
		// =======================
	}

	if rsv == "" {
		return nil, "", errors.New("Get canonical signature failed")
	}

	log.Trace(b.ChainConfig.BlockChain+" DcrmSignTransaction get rsv success", "keyID", keyID, "rsv", rsv)

	// signedTx, txHash
	signedTx, err := makeSignedTransaction([]string{rsv}, rawTx)
	eosstx, ok := signTx.(*eosgo.SignedTransaction)
	if !ok {
		return signedTx, "", fmt.Errorf("eos signed transaction type assertion error")
	}

	packedTx, err := eosstx.Pack(eosgo.CompressionNone)
	if err != nil {
		return signedTx, "", err
	}
	hashBytes, err := packedTx.ID()
	if err != nil {
		return signedTx, "", err
	}
	txHash = hashBytes.String()

	log.Info(b.ChainConfig.BlockChain+" DcrmSignTransaction success", "keyID", keyID, "txhash", txHash)
	return signedTx, txHash, err
}

// SignTransaction sign tx with pairID
func (b *Bridge) SignTransaction(rawTx interface{}, pairID string) (signTx interface{}, txHash string, err error) {
	privKey := b.GetTokenConfig(pairID).GetDcrmAddressPrivateKey()
	return b.SignTransactionWithPrivateKey(rawTx, privKey)
}

// SignTransactionWithPrivateKey sign tx with ECDSA private key
// works with uncompressed pubkey
// for test only
func (b *Bridge) SignTransactionWithPrivateKey(rawTx interface{}, privKey *ecdsa.PrivateKey) (signTx interface{}, txHash string, err error) {
	log.Infof("======\nSignTransactionWithPrivateKey\n======\n")
	eostx, ok := rawTx.(*eosgo.Transaction)
	if !ok {
		return nil, "", errors.New("raw tx type assertion error")
	}
	stx := eosgo.NewSignedTransaction(eostx)

	txdata, cfd, err := stx.PackedTransactionAndCFD()
	if err != nil {
		return nil, "", err
	}
	digest := eosgo.SigDigest(opts.ChainID, txdata, cfd)

	var sig []byte

	vrs, err := (*btcec.PrivateKey)(privKey).SignCanonical(btcec.S256(), digest)
	if err != nil {
		return nil, "", err
	}
	if vrs == nil || len(vrs) < 1 {
		return nil, "", fmt.Errorf("eos make canonical signature failed")
	}
	log.Infof("======\n111111\nvrs:\n%v\n======\n", vrs)

	//recoveredKey, _, err := btcec.RecoverCompact(btcec.S256(), vrs, digest)
	//fmt.Printf("\n\n======\nrecovered key:\n%v\n======\n\n", recoveredKey)

	sig = append(vrs[1:], vrs[0]-byte(31))

	rsv := hex.EncodeToString(sig)

	signedTx, err := makeSignedTransaction([]string{rsv}, rawTx)
	if err != nil {
		return nil, "", fmt.Errorf("sign tx failed, %v", err)
	}

	eosstx, ok := signedTx.(*eosgo.SignedTransaction)
	if !ok {
		return signedTx, "", fmt.Errorf("eos signed transaction type assertion error")
	}

	packedTx, err := eosstx.Pack(eosgo.CompressionNone)
	if err != nil {
		return signedTx, "", err
	}
	hashBytes, err := packedTx.ID()
	if err != nil {
		return signedTx, "", err
	}
	txHash = hashBytes.String()
	log.Info(b.ChainConfig.BlockChain+" SignTransaction success", "txhash", txHash)
	return signedTx, txHash, err
}

// SignedMessage
func makeSignedTransaction(rsv []string, tx interface{}) (signedTransaction interface{}, err error) {
	eostx, ok := tx.(*eosgo.Transaction)
	if !ok {
		return nil, errors.New("raw tx type assertion error")
	}

	stx := eosgo.NewSignedTransaction(eostx)

	signature, err := RSVToEOSSignature(rsv[0])
	if err != nil {
		return
	}

	stx.Signatures = append(stx.Signatures, signature)
	signedTransaction = stx
	return
}

/*
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
		//log.Error("dcrm sign sender mismath", "inputCode", args.InputCode, "have", args.From, "want", signerAddr)
		log.Error("dcrm sign sender mismath", "have", args.From, "want", signerAddr)
		return rootPubkey, fmt.Errorf("dcrm sign sender mismath")
	}
	return rootPubkey, nil
}
*/

// RSVToEOSSignature convert rsv to EOS signature
func RSVToEOSSignature(rsvStr string) (ecc.Signature, error) {
	rsv, _ := hex.DecodeString(rsvStr)
	rsv[64] += byte(31)
	v := rsv[64]
	rs := rsv[:64]
	vrs := append([]byte{v}, rs...)
	data := append([]byte{0}, vrs...)
	return ecc.NewSignatureFromData(data)
}

// IsCanonical checks if signature is canonical
func IsCanonical(compactSig []byte) bool {
	// From EOS's codebase, our way of doing Canonical sigs.
	// https://steemit.com/steem/@dantheman/steem-and-bitshares-cryptographic-security-update
	//
	// !(c.data[1] & 0x80)
	// && !(c.data[1] == 0 && !(c.data[2] & 0x80))
	// && !(c.data[33] & 0x80)
	// && !(c.data[33] == 0 && !(c.data[34] & 0x80));

	d := compactSig
	fmt.Printf("d is %v\n", d)
	t1 := (d[1] & 0x80) == 0
	t2 := !(d[1] == 0 && ((d[2] & 0x80) == 0))
	t3 := (d[33] & 0x80) == 0
	t4 := !(d[33] == 0 && ((d[34] & 0x80) == 0))

	return t1 && t2 && t3 && t4
}
