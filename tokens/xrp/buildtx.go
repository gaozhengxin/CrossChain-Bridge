package xrp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/xrp/rubblelabs/ripple/crypto"
	"github.com/anyswap/CrossChain-Bridge/tokens/xrp/rubblelabs/ripple/data"
)

var (
	defaultFee int64 = 10
)

// BuildRawTransaction build raw tx
func (b *Bridge) BuildRawTransaction(args *tokens.BuildTxArgs) (rawTx interface{}, err error) {
	var (
		sequence uint32
		fee      int64
		pubkey   string
		pairID   = args.PairID
		token    = b.GetTokenConfig(pairID)
		from     = args.From
		to       = args.To
		amount   = args.Value
	)

	if token == nil {
		return nil, fmt.Errorf("swap pair '%v' is not configed", pairID)
	}

	switch args.SwapType {
	case tokens.SwapinType:
		return nil, tokens.ErrSwapTypeNotSupported
	case tokens.SwapoutType:
		from = token.DcrmAddress                                          // from
		to = args.Bind                                                    // to
		amount = tokens.CalcSwappedValue(pairID, args.OriginValue, false) // amount
	}

	if from == "" {
		return nil, errors.New("no sender specified")
	}

	var extra *tokens.XrpExtraArgs
	if args.Extra == nil || args.Extra.XrpExtra == nil {
		extra = b.swapoutDefaultArgs()
		args.Extra = &tokens.AllExtras{XrpExtra: extra}
		sequence = *extra.Sequence
		fee = *extra.Fee
		pubkey = args.Extra.XrpExtra.FromPublic
	} else {
		extra = args.Extra.XrpExtra
		if extra.Sequence != nil {
			sequence = *extra.Sequence
		}
		if extra.Fee != nil {
			fee = *extra.Fee
		}
		pubkey = args.Extra.XrpExtra.FromPublic
	}

	if args.SwapType != tokens.NoSwapType {
		args.Identifier = params.GetIdentifier()
	}

	bal, err := b.GetBalance(from)
	if err != nil {
		log.Warn("Get from address balance error", "error", err)
	}
	remain := new(big.Int).Sub(bal, amount)
	if remain.Cmp(big.NewInt(20000000)) < 1 {
		return nil, fmt.Errorf("Insufficient xrp balance")
	}

	rawtx, _, err := b.BuildUnsignedTransaction(from, pubkey, to, amount, sequence, fee)
	return rawtx, err
}

func (b *Bridge) swapoutDefaultArgs() *tokens.XrpExtraArgs {
	args := &tokens.XrpExtraArgs{
		FromPublic: b.GetDcrmPublicKey(pairID),
		Sequence:   new(uint32),
		Fee:        new(int64),
	}

	token := b.GetTokenConfig(pairID)
	if token == nil {
		log.Warn("Swap pair id not configed", "pairID", pairID)
		return args
	}

	dcrmAddr := token.DcrmAddress

	seq, err := b.GetSeq(dcrmAddr)
	if err != nil {
		log.Warn("Get sequence error when setting default xrp args", "error", err)
	}
	*args.Sequence = seq
	addPercent := token.PlusGasPricePercentage
	if addPercent > 0 {
		*args.Fee = *args.Fee * (int64(100 + addPercent)) / 100
	}
	if *args.Fee < defaultFee {
		*args.Fee = defaultFee
	}
	return args
}

// BuildUnsignedTransaction build ripple unsigned transaction
func (b *Bridge) BuildUnsignedTransaction(fromAddress, fromPublicKey, toAddress string, amount *big.Int, sequence uint32, fee int64) (transaction interface{}, digests []string, err error) {
	pub, err := hex.DecodeString(fromPublicKey)
	xrpPubKey := ImportPublicKey(pub)
	amt := amount.String()
	txseq, err := b.GetSeq(fromAddress)
	if err != nil {
		return nil, nil, err
	}
	memo := ""
	transaction, hash, _ := NewUnsignedPaymentTransaction(xrpPubKey, nil, txseq, toAddress, amt, fee, memo, "", false, false, false)
	digests = append(digests, hash.String())
	return
}

// GetSeq returns account tx sequence
func (b *Bridge) GetSeq(address string) (uint32, error) {
	account, err := b.GetAccount(address)
	if err != nil {
		return 0, fmt.Errorf("cannot get account, %v", err)
	}
	if seq := account.AccountData.Sequence; seq != nil {
		return *seq, nil
	}
	return 0, nil // unexpected
}

// NewUnsignedPaymentTransaction build xrp payment tx
// Partial and limit must be false
func NewUnsignedPaymentTransaction(key crypto.Key, keyseq *uint32, txseq uint32, dest string, amt string, fee int64, memo string, path string, nodirect bool, partial bool, limit bool) (data.Transaction, data.Hash256, []byte) {
	if partial == true {
		log.Warn("Building tx with partial")
	}
	if limit == true {
		log.Warn("Building tx with limit")
	}

	destination, amount := parseAccount(dest), parseAmount(amt)
	payment := &data.Payment{
		Destination: *destination,
		Amount:      *amount,
	}
	payment.TransactionType = data.PAYMENT

	if memo != "" {
		memoStr := new(data.Memo)
		memoStr.Memo.MemoType = []byte("BIND")
		var memodata []byte
		if b, err := hex.DecodeString(memo); err != nil {
			memodata = []byte(memo)
		} else {
			memodata = b
		}
		memoStr.Memo.MemoData = memodata
		payment.Memos = append(payment.Memos, *memoStr)
	}

	if path != "" {
		payment.Paths = parsePaths(path)
	}
	payment.Flags = new(data.TransactionFlag)
	if nodirect {
		*payment.Flags = *payment.Flags | data.TxNoDirectRipple
	}
	if partial {
		*payment.Flags = *payment.Flags | data.TxPartialPayment
	}
	if limit {
		*payment.Flags = *payment.Flags | data.TxLimitQuality
	}

	base := payment.GetBase()

	base.Sequence = txseq

	fei, err := data.NewNativeValue(fee)
	if err != nil {
		return nil, data.Hash256{}, nil
	}
	base.Fee = *fei

	copy(base.Account[:], key.Id(keyseq))

	payment.InitialiseForSigning()
	copy(payment.GetPublicKey().Bytes(), key.Public(keyseq))
	hash, msg, err := data.SigningHash(payment)
	if err != nil {
		log.Warn("Generate ripple tx signing hash error", "error", err)
		return nil, data.Hash256{}, nil
	}
	log.Info("Build unsigned tx success", "signing hash", hash.String(), "blob", fmt.Sprintf("%X", msg))

	return payment, hash, msg
}
