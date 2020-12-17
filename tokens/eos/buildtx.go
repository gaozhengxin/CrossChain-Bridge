package eos

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"

	eosgo "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
)

var (
	retryRPCCount    = 3
	retryRPCInterval = 1 * time.Second
	opts             = &eosgo.TxOptions{
		ChainID:          hexToChecksum256(ChainID),
		MaxNetUsageWords: uint32(999),
		//DelaySecs: uint32(120),
		MaxCPUUsageMS: uint8(200),
		Compress:      eosgo.CompressionNone,
	}
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

	extra := &tokens.EOSExtraArgs{
		Memo: "AnySwap out",
	}

	args.Value = tokens.CalcSwappedValue(PairID, args.OriginValue, false)

	return b.buildTx(args, extra, input)
}

func (b *Bridge) buildTx(args *tokens.BuildTxArgs, extra *tokens.EOSExtraArgs, input []byte) (rawTx interface{}, err error) {
	cli := b.GetClient()

	var (
		value = args.Value
	)
	if value == nil {
		value = tokens.CalcSwappedValue(PairID, args.OriginValue, false)
	}
	if value == nil {
		return rawTx, fmt.Errorf("value not set")
	}
	from := args.From
	to := args.To

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
	balance = cli.GetEOSBalance(eosgo.AccountName(args.From))
	needValue := big.NewInt(0)
	if value != nil && value.Sign() > 0 {
		needValue = value
	}

	if balance.Cmp(needValue) < 0 {
		return nil, errors.New("not enough coin balance")
	}

	s := strconv.FormatFloat(float64(value.Int64())/10000, 'f', 4, 64) + " EOS"
	quantity, _ := eosgo.NewAsset(s)

	transfer := &eosgo.Action{
		Account: eosgo.AN("eosio.token"),
		Name:    eosgo.ActN("transfer"),
		Authorization: []eosgo.PermissionLevel{
			{
				Actor:      eosgo.AccountName(from),
				Permission: eosgo.PN("active"),
			},
		},
		ActionData: eosgo.NewActionData(token.Transfer{
			From:     eosgo.AccountName(from),
			To:       eosgo.AccountName(to),
			Quantity: quantity,
			Memo:     extra.Memo,
		}),
	}

	err = cli.FillFromChain(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	var actions []*eosgo.Action
	actions = append(actions, transfer)

	rawTx = eosgo.NewTransaction(actions, opts)

	log.Trace("build raw tx", "pairID", args.PairID, "identifier", args.Identifier,
		"swapID", args.SwapID, "swapType", args.SwapType,
		"bind", args.Bind, "originValue", args.OriginValue, "value", args.Value,
		"from", args.From, "to", to, "value", value, "memo", args.Extra.EOSExtra.Memo)

	return rawTx, nil
}

func hexToChecksum256(data string) eosgo.Checksum256 {
	bytes, _ := hex.DecodeString(data)
	return eosgo.Checksum256(bytes)
}
