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
)

var (
	retryRPCCount    = 3
	retryRPCInterval = 1 * time.Second
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

	extra := &token.EOSExtraArgs{
		Memo: "AnySwap out",
	}

	return b.buildTx(args, extra, input)
}

func (b *Bridge) buildTx(args *tokens.BuildTxArgs, extra *tokens.EOSExtraArgs, input []byte) (rawTx interface{}, err error) {
	cli := GetClient()

	var (
		value = args.Value
	)
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
	balance = cli.GetEOSBalance(args.From)
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
		Authorization: []eos.PermissionLevel{
			{
				Actor:      from,
				Permission: eosgo.PN("active"),
			},
		},
		ActionData: eosgo.NewActionData(token.Transfer{
			From:     from,
			To:       to,
			Quantity: quantity,
			Memo:     extra.Memo,
		}),
	}

	opts & eosgo.TxOptions{}
	err := cli.FillFromChain(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	if opts.ChainID != hexToChecksum256(ChainID) {
		return nil, fmt.Errorf("wrong chain id")
	}

	var actions []*eosgo.Action
	actions = append(actions, transfer)

	rawTx = eosgo.NewTransaction(actions, opts)

	log.Trace("build raw tx", "pairID", args.PairID, "identifier", args.Identifier,
		"swapID", args.SwapID, "swapType", args.SwapType,
		"bind", args.Bind, "originValue", args.OriginValue,
		"from", args.From, "to", to.String(), "value", value, "memo", args.Extra.Memo)

	return rawTx, nil
}

func hexToChecksum256(data string) eosgo.Checksum256 {
	bytes, _ := hex.DecodeString(data)
	return eosgo.Checksum256(bytes)
}
