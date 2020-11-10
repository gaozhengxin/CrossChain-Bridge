package filecoin

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	filAddress "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	bigger "github.com/filecoin-project/go-state-types/big"
	filApi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	cid "github.com/ipfs/go-cid"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

type Errs []error

func (errs Errs) Merge() error {
	if len(errs) == 0 {
		return nil
	}
	msgs := []string{}
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return fmt.Errorf("errors: %+v", msgs)
}

type clientGetter func() (apistruct.FullNodeStruct, func(), error)

func (b *Bridge) getClient(apiAddress string) (apistruct.FullNodeStruct, func(), error) {
	apis := b.GatewayConfig.AuthAPIs
	for _, authAPI := range apis {
		if apiAddress == authAPI.Address {
			authType := authAPI.AuthType // Bearer Basic
			authToken := authAPI.AuthToken
			var api apistruct.FullNodeStruct
			headers := http.Header{"Authorization": []string{authType + " " + authToken}}
			closer, err := jsonrpc.NewMergeClient(context.Background(), apiAddress, "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
			return api, closer, err
		}
	}
	return apistruct.FullNodeStruct{}, nil, fmt.Errorf("cannot configure auth API")
}

func (b *Bridge) getAllClientGetters() ([]clientGetter, error) {
	apis := b.GatewayConfig.AuthAPIs
	if len(apis) < 1 {
		return nil, fmt.Errorf("No authorized api in gateway config")
	}
	getters := make([]clientGetter, 0)
	for _, authAPI := range apis {
		var authType, authToken, apiAddress string
		// For lotus jsonrpc, auth type: "Bearer"
		// For infura filecoin api, auth type: "Basic"
		authType = authAPI.AuthType

		// How to get infura filecoin api auth token
		/*
			import "encoding/base64"
			authToken := base64.StdEncoding.EncodeToString([]byte(<PROJECT_ID>:<PROJECT_SECRET>))
		*/
		authToken = authAPI.AuthToken

		// For lotus jsonrpc, apiAddress = "ws://127.0.0.1:1234/rpc/vo"
		// For infura filecoin api, apiAddress = "wss://filecoin.infura.io"
		apiAddress = authAPI.Address
		getter := func() (apistruct.FullNodeStruct, func(), error) {
			headers := http.Header{"Authorization": []string{authType + " " + authToken}}
			var api apistruct.FullNodeStruct
			closer, err := jsonrpc.NewMergeClient(context.Background(), apiAddress, "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
			return api, closer, err
		}
		getters = append(getters, getter)
	}
	return getters, nil
}

// GetBalance call eth_getBalance
func (b *Bridge) GetBalance(account string) (*big.Int, error) {
	return b.getAddressBalance(account)
}

func (b *Bridge) getAddressBalance(address string) (balance *big.Int, err error) {
	addr, err := filAddress.NewFromString(address)
	if err != nil {
		return nil, err
	}
	getters, err := b.getAllClientGetters()
	if err != nil {
		return nil, err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("getAddressBalance", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()

		actor, err := api.StateGetActor(context.Background(), addr, filTypes.TipSetKey{})

		bal := bigger.Int(actor.Balance)
		bb, _ := (&bal).Bytes()
		balance = new(big.Int).SetBytes(bb)

		if err == nil {
			return balance, err
		}
		log.Trace("getAddressBalance", "error", err)
		errs = append(errs, err)
	}
	return nil, errs.Merge()
}

// GetTransactionByHash get message by cid
func (b *Bridge) GetTransactionByHash(txhash string) (filTypes.Message, error) {
	msgid, err := cid.Decode(txhash)
	if err != nil {
		return filTypes.Message{}, err
	}

	getters, err := b.getAllClientGetters()
	if err != nil {
		return filTypes.Message{}, err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("GetTransactionByHash", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()

		msg, err := api.ChainGetMessage(context.Background(), msgid)

		if err == nil {
			return *msg, nil
		}
		log.Trace("GetTransactionByHash", "error", err)
		errs = append(errs, err)
	}
	return filTypes.Message{}, errs.Merge()
}

func (b *Bridge) estimateGasLimit(arg *tokens.BuildTxArgs) int64 {
	from, _ := filAddress.NewFromString(arg.From)
	to, _ := filAddress.NewFromString(arg.To)
	msg := &filTypes.Message{
		From:   from,
		To:     to,
		Method: 0,
		Value:  abi.TokenAmount(bigger.Int{Int: arg.Value}),
	}
	getters, err := b.getAllClientGetters()
	if err != nil {
		return 1000000
	}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("estimateGasLimit", "error", err)
			continue
		}
		defer closer()
		egl, err := api.GasEstimateGasLimit(context.Background(), msg, filTypes.TipSetKey{})
		if err == nil {
			return egl
		}
		log.Trace("estimateGasLimit", "error", err)
	}
	return 1000000
}

func (b *Bridge) estimateGasPremium(from string, gasLimit int64) int64 {
	fromAddr, err := filAddress.NewFromString(from)
	if err != nil {
		return 100000
	}
	getters, err := b.getAllClientGetters()
	if err != nil {
		return 100000
	}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("estimateGasPremium", "error", err)
			continue
		}
		defer closer()
		egp, err := api.GasEstimateGasPremium(context.Background(), 3, fromAddr, gasLimit, filTypes.TipSetKey{})
		if err == nil {
			return egp.Int64()
		}
		log.Trace("estimateGasPremium", "error", err)
	}
	return 100000
}

func (b *Bridge) getAddressNonce(addr string) (nonce uint64, err error) {
	fromaddr, err := filAddress.NewFromString(addr)
	if err != nil {
		return 0, err
	}
	getters, err := b.getAllClientGetters()
	if err != nil {
		return 0, err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("getAddressNonce", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()
		actor, err := api.StateGetActor(context.Background(), fromaddr, filTypes.TipSetKey{})
		if err == nil {
			nonce = actor.Nonce
			return nonce, nil
		}
		log.Trace("getAddressNonce", "error", err)
		errs = append(errs, err)
	}
	return 1, errs.Merge()
}

// GetLatestBlockNumberOf get latest `epoch` number
func (b *Bridge) GetLatestBlockNumberOf(apiAddress string) (uint64, error) {
	api, closer, err := b.getClient(apiAddress)
	if err != nil {
		return 0, err
	}
	defer closer()
	tipset, err := api.ChainHead(context.Background())
	if err != nil {
		return 0, err
	}
	return uint64(tipset.Height()), nil
}

// GetLatestBlockNumber get latest `epoch` number
func (b *Bridge) GetLatestBlockNumber() (uint64, error) {
	getters, err := b.getAllClientGetters()
	if err != nil {
		return 0, err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("GetLatestBlockNumber", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()
		tipset, err := api.ChainHead(context.Background())
		if err == nil {
			return uint64(tipset.Height()), nil
		}
		log.Trace("GetLatestBlockNumber", "error", err)
		errs = append(errs, err)
	}
	return 0, errs.Merge()
}

func (b *Bridge) pushMessage(sm filTypes.SignedMessage) (string, error) {
	getters, err := b.getAllClientGetters()
	if err != nil {
		return "", err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("pushMessage", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()

		cid, err := api.MpoolPush(context.Background(), &sm)
		if err == nil {
			return cid.String(), nil
		}
		log.Trace("pushMessage", "error", err)
		errs = append(errs, err)
	}
	return "", errs.Merge()
}

// GetTransactionReceipt get message lookup result
func (b *Bridge) GetTransactionReceipt(txhash string) (filApi.MsgLookup, error) {
	msgid, err := cid.Decode(txhash)
	if err != nil {
		return filApi.MsgLookup{}, err
	}
	getters, err := b.getAllClientGetters()
	if err != nil {
		return filApi.MsgLookup{}, err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("GetTransactionReceipt", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()
		msgLookup, err := api.StateSearchMsg(context.Background(), msgid)
		if err == nil {
			return *msgLookup, nil
		}
		log.Trace("GetTransactionReceipt", "error", err)
		errs = append(errs, err)
	}
	return filApi.MsgLookup{}, errs.Merge()
}

// StartListenMpool start listen mempool updates
func (b *Bridge) StartListenMpool(ctx context.Context, callback func(filApi.MpoolUpdate)) error {
	getters, err := b.getAllClientGetters()
	if err != nil {
		return err
	}
	var sub <-chan filApi.MpoolUpdate
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("StartListenMpool", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()

		sub, err = api.MpoolSub(ctx)
		if err != nil {
			log.Trace("StartListenMpool", "error", err)
			errs = append(errs, err)
			continue
		}
	}

	if sub == nil {
		return fmt.Errorf("subscribe mempool updates faild: %+v", errs.Merge().Error())
	}

	for {
		fmt.Println("Listening")
		select {
		case mup, ok := <-sub:
			if !ok {
				return fmt.Errorf("connection with lotus node broke")
			}
			callback(mup)
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

// GetTipsetByNumber get tipset by number
func (b *Bridge) GetTipsetByNumber(height uint64) (*filTypes.TipSet, error) {
	getters, err := b.getAllClientGetters()
	if err != nil {
		return nil, err
	}
	errs := Errs{}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("GetTipsetByNumber", "error", err)
			errs = append(errs, err)
			continue
		}
		defer closer()

		tipset, err := api.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(int64(height)), filTypes.TipSetKey{})
		if err == nil {
			return tipset, nil
		}
		log.Trace("GetTipsetByNumber", "error", err)
		errs = append(errs, err)
	}
	return nil, errs.Merge()
}

// GetBlockMessages get messages from block (rather than tipset)
func (b *Bridge) GetBlockMessages(block string) (msgids []string) {
	msgids = []string{}
	blockCid, err := cid.Decode(block)
	if err != nil {
		return
	}
	getters, err := b.getAllClientGetters()
	if err != nil {
		return
	}
	for _, getter := range getters {
		var err error
		api, closer, err := getter()
		if err != nil {
			log.Trace("GetBlockMessages", "error", err)
			continue
		}
		defer closer()

		bms, err := api.ChainGetBlockMessages(context.Background(), blockCid)
		if err == nil {
			for _, bm := range bms.Cids {
				msgids = append(msgids, bm.String())
			}
			return
		}
		log.Trace("GetBlockMessages", "error", err)
	}
	return
}

// GetTokenBalance impl
func (b *Bridge) GetTokenBalance(tokenType, tokenAddress, accountAddress string) (*big.Int, error) {
	return nil, fmt.Errorf("[%v] can not get token balance of token with type '%v'", b.ChainConfig.BlockChain, tokenType)
}

// GetTokenSupply impl
func (b *Bridge) GetTokenSupply(tokenType, tokenAddress string) (*big.Int, error) {
	return nil, fmt.Errorf("[%v] can not get token supply of token with type '%v'", b.ChainConfig.BlockChain, tokenType)
}
