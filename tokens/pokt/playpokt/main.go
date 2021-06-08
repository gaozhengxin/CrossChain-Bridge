package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"github.com/pokt-network/pocket-core/codec"
	types2 "github.com/pokt-network/pocket-core/codec/types"
	"github.com/pokt-network/pocket-core/crypto"
	sdk "github.com/pokt-network/pocket-core/types"
	"github.com/pokt-network/pocket-core/types/module"
	apps "github.com/pokt-network/pocket-core/x/apps"
	"github.com/pokt-network/pocket-core/x/auth"
	"github.com/pokt-network/pocket-core/x/gov"
	"github.com/pokt-network/pocket-core/x/nodes"
	pocket "github.com/pokt-network/pocket-core/x/pocketcore"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var (
	memCDC *codec.Codec
	memCLI *Client
)

type Client struct {
	*http.Client
	Endpoint string
}

const (
	endpoint = "https://node1.testnet.pokt.network"
)

func main() {
	fmt.Printf("==========\n")

	number, err := getClient().QueryHeight()
	if err != nil {
		spew.Printf("%#v\n", err)
	}
	spew.Printf("%#v\n", number)

	fmt.Printf("==========\n")

	blk, err := getClient().QueryBlock(1000)
	if err != nil {
		spew.Printf("%#v\n", err)
	}
	spew.Printf("%#v\n", blk)

	fmt.Printf("==========\n")
}

func memCodec() *codec.Codec {
	if memCDC == nil {
		memCDC = codec.NewCodec(types2.NewInterfaceRegistry())
		module.NewBasicManager(
			apps.AppModuleBasic{},
			auth.AppModuleBasic{},
			gov.AppModuleBasic{},
			nodes.AppModuleBasic{},
			pocket.AppModuleBasic{},
		).RegisterCodec(memCDC)
		sdk.RegisterCodec(memCDC)
		crypto.RegisterAmino(memCDC.AminoCodec().Amino)
	}
	return memCDC
}

func (c *Client) QueryHeight() (uint64, error) {
	var data = strings.NewReader(`{}`)
	req, err := http.NewRequest("POST", c.Endpoint+"/v1/query/height", data)
	if err != nil {
		return 0, errors.Wrap(err, "QueryHeight")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "QueryHeight")
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "QueryHeight")
	}
	var height struct {
		Height int64
	}
	err = json.Unmarshal(bodyText, &height)
	if err != nil {
		return 0, errors.Wrap(err, "QueryHeight")
	}
	return uint64(height.Height), nil
}

func (c *Client) QueryBlock(height int64) (*ctypes.ResultBlock, error) {
	var data = strings.NewReader(spew.Sprintf("{\"height\":%v}", height))
	req, err := http.NewRequest("POST", c.Endpoint+"/v1/query/block", data)
	if err != nil {
		return nil, errors.Wrap(err, "QueryBlock")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "QueryBlock")
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "QueryBlock")
	}
	var blk ctypes.ResultBlock
	err = memCodec().UnmarshalJSON(bodyText, &blk)
	if err != nil {
		return nil, errors.Wrap(err, "QueryBlock")
	}
	return &blk, nil
}

func getClient() *Client {
	if memCLI == nil {
		memCLI = &Client{
			Client:   &http.Client{},
			Endpoint: endpoint,
		}
	}
	return memCLI
}
