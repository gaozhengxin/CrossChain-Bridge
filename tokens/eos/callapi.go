package eos

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/anyswap/CrossChain-Bridge/log"
	eosgo "github.com/eoscanada/eos-go"
)

var cli *Client

var (
	// EOSAPITimeout EOS api timeout
	EOSAPITimeout = time.Second * 120
	// EOSAPILongTimeout EOS api long timeout
	EOSAPILongTimeout = time.Second * 600
	// EOSAPIRetryTimes is EOS call api retry times
	EOSAPIRetryTimes = 20
	// EOSAPIRetryInterval is EOS call api retry time interval
	EOSAPIRetryInterval = time.Millisecond * 100

	// ErrAPITimeout api timeout error
	ErrAPITimeout = fmt.Errorf("EOS call api timeout")
	// ErrAPIFail api call fail error
	ErrAPIFail = fmt.Errorf("EOS call api fail")
)

// NewClient returns new client
func (b *Bridge) NewClient() *Client {
	cli := &Client{
		APIs: make(map[string](*eosgo.API)),
	}
	for _, apiaddr := range b.GatewayConfig.APIAddress {
		cli.APIs[apiaddr] = eosgo.New(apiaddr)
	}
	return cli
}

// NewClient returns new client
func NewClient(apiaddr string) *Client {
	cli := &Client{
		APIs: make(map[string](*eosgo.API)),
	}
	cli.APIs[apiaddr] = eosgo.New(apiaddr)
	return cli
}

// GetClient returns a client
func (b *Bridge) GetClient() *Client {
	if cli != nil {
		return cli
	}
	cli = b.NewClient()
	return cli
}

// Client is a type of wrapped EOS api client
type Client struct {
	APIs map[string](*eosgo.API)
}

// single eos.API has timeout 30s by default
func (cli *Client) getAPI(addr string) *eosgo.API {
	if cli.APIs[addr] == nil {
		cli.APIs[addr] = eosgo.New(addr)
	}
	return cli.APIs[addr]
}

func (cli *Client) callAPI(ctx context.Context, do func(ctx context.Context, api *eosgo.API, resch chan (interface{}))) (resp interface{}, err error) {
	for i := 0; i < EOSAPIRetryTimes; i++ {
		resp, err = cli.callAPIOnce(ctx, do)
		if resp != nil {
			return resp, err
		}
		log.Debug("EOS call api fail", "error", err)
		time.Sleep(EOSAPIRetryInterval)
	}
	return
}

func (cli *Client) callAPIOnce(ctx context.Context, do func(ctx context.Context, api *eosgo.API, resch chan (interface{}))) (resp interface{}, err error) {

	resch := make(chan interface{}, 1)

	var wg sync.WaitGroup

	for addr, api := range cli.APIs {
		apiConn := api
		if apiConn == nil {
			apiConn = cli.getAPI(addr)
		}
		wg.Add(1)
		go func() {
			do(ctx, apiConn, resch)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(resch)
	}()

	select {
	case resp, ok := <-resch:
		if ok && resp != nil {
			return resp, nil
		}
	case <-ctx.Done():
		return nil, ErrAPITimeout
	}
	return nil, ErrAPIFail
}

func checkPanic() {
	if r := recover(); r != nil {
		return
	}
}

// GetAccount gets account info
func (cli *Client) GetAccount(ctx context.Context, name eosgo.AccountName) (out *eosgo.AccountResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.GetAccount(name)
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.AccountResp); ok {
		return out, err
	}
	return nil, err
}

// GetInfo gets chain info
func (cli *Client) GetInfo(ctx context.Context) (out *eosgo.InfoResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.GetInfo()
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.InfoResp); ok {
		return out, err
	}
	return nil, err
}

// GetCurrencyBalance gets currency balance
func (cli *Client) GetCurrencyBalance(ctx context.Context, account eosgo.AccountName, symbol string, code eosgo.AccountName) (out []eosgo.Asset, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.GetCurrencyBalance(account, symbol, code)
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.([]eosgo.Asset); ok {
		return out, err
	}
	return nil, err
}

// GetEOSBalance gets EOS balance
func (cli *Client) GetEOSBalance(account eosgo.AccountName) *big.Int {
	asts, err := cli.GetCurrencyBalance(context.Background(), account, "EOS", "eosio.token")
	if len(asts) == 0 || err != nil {
		return big.NewInt(0)
	}
	for _, ast := range asts {
		if ast.Symbol.Symbol == "EOS" {
			return big.NewInt(int64(ast.Amount))
		}
	}
	return big.NewInt(0)
}

// FillFromChain auto fills transaction options
func (cli *Client) FillFromChain(ctx context.Context, opts *eosgo.TxOptions) error {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		err := opts.FillFromChain(api)
		if err == nil {
			resch <- opts
		}
	})
	if _, ok := resp.(*eosgo.TxOptions); ok {
		return nil
	}
	return err
}

// PushTransaction pushes transaction
func (cli *Client) PushTransaction(ctx context.Context, tx *eosgo.PackedTransaction) (out *eosgo.PushTransactionFullResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.PushTransaction(tx)
		fmt.Printf("\n\n======\npush transaction\nres: %+v\nerr: %+v\n======\n\n", out, err)
		if err != nil {
			log.Info("EOS PushTransaction", "error", err)
		}
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.PushTransactionFullResp); ok {
		return out, nil
	}
	return nil, err
}

// GetTransaction gets transaction by id
func (cli *Client) GetTransaction(ctx context.Context, id string) (out *eosgo.TransactionResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.GetTransaction(id)
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.TransactionResp); ok {
		if out == nil || out.Transaction.Transaction.Transaction == nil {
			return nil, fmt.Errorf("Get transaction failed")
		}
		return out, nil
	}
	return nil, err
}

// GetBlockByNum gets block by number
func (cli *Client) GetBlockByNum(ctx context.Context, num uint32) (out *eosgo.BlockResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPILongTimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.GetBlockByNum(num)
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.BlockResp); ok {
		return out, nil
	}
	return nil, err
}

// GetActions gets actions
func (cli *Client) GetActions(ctx context.Context, accountName string, pos, offset int64) (actions *eosgo.ActionsResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPILongTimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		params := eosgo.GetActionsRequest{
			AccountName: eosgo.AccountName(accountName),
			Pos:         eosgo.Int64(pos),
			Offset:      eosgo.Int64(offset),
		}
		out, err := api.GetActions(params)
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.ActionsResp); ok {
		return out, nil
	}
	return nil, err
}
