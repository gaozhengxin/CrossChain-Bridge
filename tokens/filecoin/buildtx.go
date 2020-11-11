package filecoin

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"

	filAddress "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	bigger "github.com/filecoin-project/go-state-types/big"
	filTypes "github.com/filecoin-project/lotus/chain/types"
)

var (
	retryRPCCount    = 3
	retryRPCInterval = 1 * time.Second

	defReserveGasFee = big.NewInt(1e16) // 0.01 ETH
)

// BuildRawTransaction build raw tx
func (b *Bridge) BuildRawTransaction(args *tokens.BuildTxArgs) (rawTx interface{}, err error) {
	var input []byte
	var tokenCfg *tokens.TokenConfig
	if args.Input == nil {
		if args.SwapType != tokens.NoSwapType {
			pairID := args.PairID
			tokenCfg = b.GetTokenConfig(pairID)
			if tokenCfg == nil {
				return nil, tokens.ErrUnknownPairID
			}
			if args.From == "" {
				args.From = tokenCfg.DcrmAddress // from
			}
		}
		switch args.SwapType {
		case tokens.SwapinType:
			if b.IsSrc {
				return nil, tokens.ErrBuildSwapTxInWrongEndpoint
			}
			return nil, tokens.ErrSwapTypeNotSupported
		case tokens.SwapoutType:
			if !b.IsSrc {
				return nil, tokens.ErrBuildSwapTxInWrongEndpoint
			}
			args.To = args.Bind
			input = []byte(tokens.UnlockMemoPrefix + args.SwapID)
		}
	} else {
		input = *args.Input
		if args.SwapType != tokens.NoSwapType {
			return nil, fmt.Errorf("forbid build raw swap tx with input data")
		}
	}

	extra, err := b.setDefaults(args)
	if err != nil {
		return nil, err
	}

	return b.buildTx(args, extra, input)
}

func (b *Bridge) buildTx(args *tokens.BuildTxArgs, extra *tokens.FilExtraArgs, input []byte) (rawTx interface{}, err error) {
	var (
		value     = args.Value
		nonce     = *extra.Nonce
		gasLimit  = *extra.GasLimit
		gasFeeCap = extra.GasFeeCap
	)
	from, err := filAddress.NewFromString(args.From)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %v", err)
	}
	to, err := filAddress.NewFromString(args.To)
	if err != nil {
		return nil, fmt.Errorf("invalid to address: %v", err)
	}

	if args.SwapType == tokens.SwapoutType {
		pairID := args.PairID
		tokenCfg := b.GetTokenConfig(pairID)
		if tokenCfg == nil {
			return nil, tokens.ErrUnknownPairID
		}
	}

	if args.SwapType != tokens.NoSwapType {
		args.Identifier = params.GetIdentifier()
	}

	var balance *big.Int
	balance, err = b.getAddressBalance(args.From)
	if err != nil {
		log.Warn("get balance error", "from", args.From, "err", err)
		return nil, fmt.Errorf("get balance error: %v", err)
	}
	needValue := big.NewInt(0)
	if value != nil && value.Sign() > 0 {
		needValue = value
	}
	gasFee := new(big.Int).Mul(big.NewInt(gasLimit), gasFeeCap)
	needValue = new(big.Int).Add(needValue, gasFee)

	if balance.Cmp(needValue) < 0 {
		return nil, errors.New("not enough coin balance")
	}

	rawTx = &filTypes.Message{
		From:      from,
		To:        to,
		Method:    0,
		Value:     abi.TokenAmount(bigger.Int{Int: args.Value}),
		Nonce:     nonce,
		GasLimit:  gasLimit,
		GasFeeCap: bigger.Int{Int: gasFeeCap},
	}

	log.Trace("build raw tx", "pairID", args.PairID, "identifier", args.Identifier,
		"swapID", args.SwapID, "swapType", args.SwapType,
		"bind", args.Bind, "originValue", args.OriginValue,
		"from", args.From, "to", to.String(), "value", value, "nonce", nonce,
		"gasLimit", gasLimit, "gasFeeCap", gasFeeCap, "data", common.ToHex(input))

	return rawTx, nil
}

func (b *Bridge) setDefaults(args *tokens.BuildTxArgs) (extra *tokens.FilExtraArgs, err error) {
	if args.Value == nil {
		args.Value = new(big.Int)
	}
	if args.Extra == nil || args.Extra.FilExtra == nil {
		extra = &tokens.FilExtraArgs{}
		args.Extra = &tokens.AllExtras{FilExtra: extra}
	} else {
		extra = args.Extra.FilExtra
	}
	if extra.GasLimit == nil {
		egl := b.estimateGasLimit(args)
		gasLimit := int64(egl + 1000)
		extra.GasLimit = &gasLimit
	}
	if extra.GasFeeCap == nil {
		egp := b.estimateGasPremium(args.From, *extra.GasLimit)
		extra.GasFeeCap = big.NewInt(egp + 1000)
	}
	if extra.Nonce == nil {
		extra.Nonce, err = b.getAccountNonce(args.PairID, args.From, args.SwapType)
		if err != nil {
			return nil, err
		}
	}
	return extra, nil
}

func (b *Bridge) getAccountNonce(pairID, from string, swapType tokens.SwapType) (nonceptr *uint64, err error) {
	var nonce uint64
	nonce, err = b.getAddressNonce(from)
	if err != nil {
		return nil, err
	}
	if swapType != tokens.NoSwapType {
		tokenCfg := b.GetTokenConfig(pairID)
		if tokenCfg != nil && from == tokenCfg.DcrmAddress {
			nonce = b.AdjustNonce(pairID, nonce)
		}
	}
	return &nonce, nil
}
