package eos

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
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
	gettx, err := b.GetTransaction(txHash)
	if err != nil {
		log.Debug(b.ChainConfig.BlockChain+" Bridge::GetTransaction fail", "tx", txHash, "err", err)
		return swapInfo, err
	}

	txresp, ok := gettx.(*eosgo.TransactionResp)
	if !ok {
		return swapInfo, fmt.Errorf("Get eos transaction type assertion error")
	}

	tx := txresp.Transaction.Transaction

	var from string
	var txRecipient string
	var memo string
	value := new(big.Int)
	// Check actions, match "transfer"
	for _, action := range tx.Actions {
		if action.Name == eosgo.ActN("transfer") == false {
			// from
			from = string(action.Account)

			data, ok := action.ActionData.Data.(token.Transfer)
			if !ok {
				log.Warn("eos verifySwapinTx not a transfer action")
				continue
			}

			// txRecipient
			txRecipient = string(data.To)

			// transfer value
			if strings.EqualFold(data.Quantity.Symbol.Symbol, "EOS") == false {
				continue
			}
			value = big.NewInt(int64(data.Quantity.Amount))

			// memo
			memo = data.Memo
			break
		}
	}

	token := b.GetTokenConfig(pairID)
	if token == nil {
		return nil, tokens.ErrUnknownPairID
	}

	if !common.IsEqualIgnoreCase(txRecipient, token.DepositAddress) {
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

	bindAddress, err := GetBindAddress(swapInfo.From, swapInfo.To, token.DepositAddress, memo, pairCfg)
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
	if pairCfg.UseBip32 {
		return "", fmt.Errorf("EOS not support BIP32")
	}
	if !common.IsEqualIgnoreCase(to, depositAddress) {
		return "", tokens.ErrTxWithWrongReceiver
	}
	if !tools.IsAddressRegistered(from, pairCfg) {
		return "", tokens.ErrTxSenderNotRegistered
	}
	bindAddr, err := getBindAddressFromMemo(memo)
	if err != nil {
		return "", err
	}
	return bindAddr, nil
}

func getBindAddressFromMemo(memo string) (string, error) {
	bindAddress := memo
	if !tokens.DstBridge.IsValidAddress(bindAddress) {
		return "", fmt.Errorf("wrong memo bind address")
	}
	return bindAddress, nil
}

func (b *Bridge) getStableReceipt(swapInfo *tokens.TxSwapInfo) error {
	txStatus := b.GetTransactionStatus(swapInfo.Hash)
	swapInfo.Height = txStatus.BlockHeight // Height
	confirmations := *b.GetChainConfig().Confirmations
	if txStatus.BlockHeight > 0 && txStatus.Confirmations >= confirmations {
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
