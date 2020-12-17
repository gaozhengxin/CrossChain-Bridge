package eos

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"

	eosgo "github.com/eoscanada/eos-go"
)

// GetTransaction impl
func (b *Bridge) GetTransaction(txHash string) (interface{}, error) {
	cli := b.GetClient()
	return cli.GetTransaction(context.Background(), txHash)
}

// GetTransactionStatus impl
func (b *Bridge) GetTransactionStatus(txHash string) *tokens.TxStatus {
	cli := b.GetClient()

	tx, err := cli.GetTransaction(context.Background(), txHash)

	if err != nil {
		return nil
	}

	if tx == nil {
		return nil
	}

	// must excuted
	if tx.Receipt.Status == (eosgo.TransactionStatusExecuted) == false {
		return nil
	}

	var txStatus tokens.TxStatus

	txStatus.BlockHeight = uint64(tx.BlockNum)

	current := uint64(tx.LastIrreversibleBlock)

	txStatus.Confirmations = current - txStatus.BlockHeight

	txStatus.Receipt = tx.Receipt
	return &txStatus
}

// VerifyMsgHash verify msg hash
func (b *Bridge) VerifyMsgHash(rawTx interface{}, msgHashes []string) error {
	tx, ok := rawTx.(*eosgo.Transaction)
	if !ok {
		return tokens.ErrWrongRawTx
	}

	if len(msgHashes) != 1 {
		return tokens.ErrWrongCountOfMsgHashes
	}

	stx := eosgo.NewSignedTransaction(tx)

	txdata, cfd, err := stx.PackedTransactionAndCFD()
	if err != nil {
		return err
	}

	sigHash := eosgo.SigDigest(opts.ChainID, txdata, cfd)
	sigHashStr := hex.EncodeToString(sigHash)

	if strings.EqualFold(msgHashes[0], sigHashStr) == false {
		log.Trace("message hash mismatch", "want", msgHashes[0], "have", sigHashStr)
		return tokens.ErrMsgHashMismatch
	}
	return nil
}

// VerifyTransaction impl
func (b *Bridge) VerifyTransaction(pairID, txHash string, allowUnstable bool) (*tokens.TxSwapInfo, error) {
	if !b.IsSrc {
		return nil, tokens.ErrBridgeDestinationNotSupported
	}
	return b.verifySwapinTx(pairID, txHash, allowUnstable)
}

// verifySwapinTx verify swapin (in scan job)
func (b *Bridge) verifySwapinTx(pairID, txHash string, allowUnstable bool) (swapInfo *tokens.TxSwapInfo, err error) {
	fmt.Printf("\n======\nverifySwapinTx\n======\n")
	var txresp *eosgo.TransactionResp
	for i := 0; i < 10; i++ {
		gettx, err := b.GetTransaction(txHash)
		if err != nil {
			log.Debug(b.ChainConfig.BlockChain+" Bridge::GetTransaction fail", "tx", txHash, "err", err)
			time.Sleep(time.Second * 1)
			continue
		}

		resp, ok := gettx.(*eosgo.TransactionResp)
		if !ok {
			err = fmt.Errorf("Get eos transaction type assertion error")
			time.Sleep(time.Second * 1)
			continue
		}

		txresp = resp

		if &txresp.Receipt == nil {
			err = fmt.Errorf("EOS swapin tx status not found")
			time.Sleep(time.Second * 1)
			continue
		}
		if txresp.Receipt.Status != eosgo.TransactionStatusExecuted {
			err = fmt.Errorf("EOS swapin tx status has wrong status: %v", txresp.Receipt.Status)
			time.Sleep(time.Second * 1)
			continue
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		return swapInfo, err
	}
	if txresp == nil {
		return swapInfo, fmt.Errorf("not supposed error, txresp is nil")
	}

	fmt.Printf("\n======\n000000\n%+v\n======\n", txresp.Transaction)
	tx := txresp.Transaction.Transaction

	var from string
	var txRecipient string
	var memo string
	value := new(big.Int)
	// Check actions, match "transfer"
	for _, action := range tx.Actions {
		if action.Name == eosgo.ActN("transfer") {
			// from
			from = string(action.Account)

			data, ok := action.ActionData.Data.(map[string]interface{})
			if !ok {
				log.Warn("eos verifySwapinTx, action data error", "type", reflect.TypeOf(action.ActionData.Data), "data", action.ActionData.Data)
				continue
			}

			// txRecipient
			txRecipient, _ = data["to"].(string)

			// transfer value
			if qttstr, ok := data["quantity"].(string); ok {
				if qtt, err := eosgo.NewEOSAssetFromString(qttstr); err == nil {
					if strings.EqualFold(qtt.Symbol.Symbol, "EOS") {
						value = big.NewInt(int64(qtt.Amount))
					} else {
						log.Warn("Transfer token type is not EOS", "token", qtt.Symbol)
						continue
					}
				} else {
					log.Warn("Invalid transfer asset", "asset", qttstr)
					continue
				}
			} else {
				log.Warn("not supposed error, quantity type assertion error", "type", reflect.TypeOf(data["quantity"]), "is nil", (data["quantity"] == nil))
				continue
			}

			// memo
			memo, _ = data["memo"].(string)
			break
		}
	}
	token := b.GetTokenConfig(pairID)
	if token == nil {
		fmt.Printf("\n======\n111111\n%v\n======\n", tokens.ErrUnknownPairID)
		return nil, tokens.ErrUnknownPairID
	}

	if common.IsEqualIgnoreCase(txRecipient, token.DepositAddress) {
		fmt.Printf("\n======\n222222\n%v\n======\n", txRecipient)
		return nil, nil
	}
	pairCfg := tokens.GetTokenPairConfig(pairID)

	swapInfo = &tokens.TxSwapInfo{}
	swapInfo.Hash = txHash      // Hash
	swapInfo.PairID = pairID    // PairID
	swapInfo.TxTo = txRecipient // TxTo
	swapInfo.To = txRecipient   // To
	swapInfo.From = from        // From
	swapInfo.Value = value      // Value

	//bindAddress, err := GetBindAddress(swapInfo.From, swapInfo.To, token.DepositAddress, memo, pairCfg)
	bindAddress, err := GetBindAddress(swapInfo.From, swapInfo.To, "", memo, pairCfg)
	if err != nil {
		return swapInfo, err
	}

	swapInfo.Bind = bindAddress // Bind

	/*if !allowUnstable {
		if err = b.5(swapInfo); err != nil {
			return nil, nil
		}
	}*/
	// must get stable receipt
	if err = b.getStableReceipt(swapInfo); err != nil {
		fmt.Printf("\n======\n333333\n======\n")
		return nil, nil
	}

	err = b.checkSwapinInfo(swapInfo)

	if !allowUnstable && err == nil {
		log.Debug("verify swapin stable pass", "pairID", swapInfo.PairID, "from", swapInfo.From, "to", swapInfo.To, "bind", swapInfo.Bind, "value", swapInfo.Value, "txid", txHash, "height", swapInfo.Height, "timestamp", swapInfo.Timestamp)
	}

	return swapInfo, err
}

// GetBindAddress get bind address
func GetBindAddress(from, to, depositAddress, memo string, pairCfg *tokens.TokenPairConfig) (string, error) {
	bindAddr, err := getBindAddressFromMemo(memo)
	if err != nil {
		return "", err
	}
	return bindAddr, nil
}

func getBindAddressFromMemo(memo string) (string, error) {
	bindAddress := memo
	if !tokens.DstBridge.IsValidAddress(bindAddress) {
		return "", fmt.Errorf("wrong memo bind address, bind address: %+v", bindAddress)
	}
	return bindAddress, nil
}

func (b *Bridge) getStableReceipt(swapInfo *tokens.TxSwapInfo) error {
	txStatus := b.GetTransactionStatus(swapInfo.Hash)
	swapInfo.Height = txStatus.BlockHeight // Height
	confirmations := *b.GetChainConfig().Confirmations
	irr, _ := b.GetIrreversible()
	if txStatus.BlockHeight > irr && txStatus.Confirmations >= confirmations {
		return nil
	}
	return tokens.ErrTxWithWrongReceipt
}

func (b *Bridge) checkSwapinInfo(swapInfo *tokens.TxSwapInfo) error {
	if swapInfo.Bind == swapInfo.To {
		return tokens.ErrTxWithWrongSender
	}
	if !tokens.CheckSwapValue(swapInfo.PairID, swapInfo.Value, b.IsSrc) {
		return tokens.ErrTxWithWrongValue
	}
	return b.checkSwapinBindAddress(swapInfo.Bind)
}

func (b *Bridge) checkSwapinBindAddress(bindAddr string) error {
	if !tokens.DstBridge.IsValidAddress(bindAddr) {
		log.Warn("wrong bind address in swapin", "bind", bindAddr)
		return tokens.ErrTxWithWrongMemo
	}
	return nil
}
