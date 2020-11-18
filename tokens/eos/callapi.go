package eos

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	eosgo "github.com/eoscanada/eos-go"
)

var c *Client

var (
	// EOSAPITimeout EOS api timeout
	EOSAPITimeout = time.Second * 5

	// ErrAPITimeout api timeout error
	ErrAPITimeout = fmt.Errorf("EOS call api timeout")
	// ErrAPIFail api call fail error
	ErrAPIFail = fmt.Errorf("EOS call api fail")
)

// NewClient returns new client
func NewClient() *Client {
	return &Client{
		APIs: make(map[string](*eosgo.API)),
	}
}

// GetClient returns a client
func GetClient() *Client {
	if cli != nil {
		return r
	}
	cli = NewClient()
	return cli
}

// Client is a type of wrapped EOS api client
type Client struct {
	APIs map[string](*eosgo.API)
}

func (cli *Client) getAPI(addr string) *eosgo.API {
	if cli.APIs[addr] == nil {
		cli.APIs[addr] = eosgo.New(addr)
	}
	return cli.APIs[addr]
}

func (cli *Client) callAPI(ctx context.Context, do func(ctx context.Context, api *eosgo.API, resch chan (interface{}))) (resp interface{}, err error) {

	resch := make(chan *APIResult, 1)

	var wg sync.WaitGroup

	for _, addr := range b.GatewayConfig.APIAddresses {
		wg.Add(1)
		api := getAPI(addr)
		go func() {
			do(ctx, api, resch)
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
		return nil, ErrAPIFail
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
func (cli *Client) GetAccount(ctx context.Context, name AccountName) (out *eosgo.AccountResp, err error) {
	ctx, cancel := context.WithTimeout(ctx, EOSAPITimeout)
	defer cancel()
	resp, err := cli.callAPI(ctx, func(ctx context.Context, api *eosgo.API, resch chan (interface{})) {
		defer checkPanic()

		out, err := api.GetAccount(ctx, name)
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

		out, err := api.GetInfo(ctx)
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

		out, err := api.GetCurrencyBalance(ctx, account, symbol, code)
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
		if ast.Symbol == eosgo.Symbol("EOS") {
			return big.NewInt(ast.Amount)
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

		err := opts.FillFromChain(ctx, api)
		if err == nil && out != nil {
			resch <- opts
		}
	})
	if out, ok := resp.(*eosgo.TxOptions); ok {
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

		out, err := api.PushTransaction(ctx, tx)
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

		out, err := api.GetTransaction(ctx, tx)
		if err == nil && out != nil {
			resch <- out
		}
	})
	if out, ok := resp.(*eosgo.TransactionResp); ok {
		return out, nil
	}
	return nil, err
}
