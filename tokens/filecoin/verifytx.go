package filecoin

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"

	filAddress "github.com/filecoin-project/go-address"
	bigger "github.com/filecoin-project/go-state-types/big"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/minio/blake2b-simd"
)

// GetTransaction impl
func (b *Bridge) GetTransaction(txHash string) (interface{}, error) {
	return b.GetTransactionByHash(txHash)
}

// GetTransactionStatus impl
func (b *Bridge) GetTransactionStatus(txHash string) *tokens.TxStatus {
	var txStatus tokens.TxStatus

	msg, err := b.GetTransactionByHash(txHash)
	if err != nil {
		log.Trace("GetTransaction fail", "hash", txHash, "err", err)
		return &txStatus
	}
	msgLookup, err := b.GetTransactionReceipt(txHash)
	if err != nil {
		log.Trace("GetTransactionReceipt fail", "hash", txHash, "err", err)
		return &txStatus
	}

	// Method must be 0
	if msg.Method != 0 {
		return nil
	}

	// ExitCode must be 0
	if msgLookup.Receipt.ExitCode.IsSuccess() == false {
		return nil
	}

	txStatus.BlockHeight = uint64(msgLookup.Height)

	current, err := b.GetLatestBlockNumber()
	txStatus.Confirmations = current - txStatus.BlockHeight

	txStatus.Receipt = msgLookup
	return &txStatus
}

// VerifyMsgHash verify msg hash
func (b *Bridge) VerifyMsgHash(rawTx interface{}, msgHashes []string) error {
	msg, ok := rawTx.(*filTypes.Message)
	if !ok {
		return tokens.ErrWrongRawTx
	}
	if len(msgHashes) != 1 {
		return tokens.ErrWrongCountOfMsgHashes
	}
	msgHash := msgHashes[0]

	mb, err := msg.ToStorageBlock()
	if err != nil {
		return fmt.Errorf("filecoin message to storage block error: %v", err)
	}

	msgBytes := mb.Cid().Bytes()
	bb := blake2b.Sum256(msgBytes)
	sigHash := hex.EncodeToString(bb[:])

	if sigHash != msgHash {
		log.Trace("message hash mismatch", "want", msgHash, "have", sigHash)
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
	tx, err := b.GetTransactionByHash(txHash)
	if err != nil {
		log.Debug(b.ChainConfig.BlockChain+" Bridge::GetTransaction fail", "tx", txHash, "err", err)
		return swapInfo, err
	}
	if tx.To == (filAddress.Address{}) { // ignore contract creation tx
		return swapInfo, err
	}
	txRecipient := strings.ToLower(tx.To.String())
	token := b.GetTokenConfig(pairID)
	if token == nil {
		return nil, tokens.ErrUnknownPairID
	}

	if !common.IsEqualIgnoreCase(txRecipient, token.DepositAddress) {
		return nil, nil
	}

	v := bigger.Int(tx.Value)
	vb, _ := (&v).Bytes()

	swapInfo = &tokens.TxSwapInfo{}
	swapInfo.Hash = txHash                            // Hash
	swapInfo.PairID = pairID                          // PairID
	swapInfo.TxTo = txRecipient                       // TxTo
	swapInfo.To = txRecipient                         // To
	swapInfo.From = strings.ToLower(tx.From.String()) // From
	swapInfo.Bind = swapInfo.From                     // Bind
	swapInfo.Value = new(big.Int).SetBytes(vb)        // Value

	/*if !allowUnstable {
		if err = b.getStableReceipt(swapInfo); err != nil {
			return nil, nil
		}
	}*/
	// always get stable receipt
	if err = b.getStableReceipt(swapInfo); err != nil {
		return nil, nil
	}

	err = b.checkSwapinInfo(swapInfo)

	if !allowUnstable && err == nil {
		log.Debug("verify swapin stable pass", "pairID", swapInfo.PairID, "from", swapInfo.From, "to", swapInfo.To, "bind", swapInfo.Bind, "value", swapInfo.Value, "txid", txHash, "height", swapInfo.Height, "timestamp", swapInfo.Timestamp)
	}

	return swapInfo, err
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
	if !tools.IsAddressRegistered(bindAddr) {
		return tokens.ErrTxSenderNotRegistered
	}
	return nil
}
