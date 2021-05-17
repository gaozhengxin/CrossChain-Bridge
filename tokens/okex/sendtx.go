package okex

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/tools/rlp"
	"github.com/anyswap/CrossChain-Bridge/types"
	"github.com/anyswap/CrossChain-Bridge/common"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	tx, ok := signedTx.(*types.Transaction)
	if !ok {
		fmt.Printf("signed tx is %+v\n", signedTx)
		return "", errors.New("wrong signed transaction type")
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return "", err
	}

	hexData := common.ToHex(data)

	gateway := b.GatewayConfig

	errs := make([]error, 0)
	for _, apiAddress := range gateway.APIAddress {
		txhash ,err := postRawTx(apiAddress, hexData)
		if err == nil {
			return txhash, nil
		}
		errs = append(errs, err)
	}
	return "", fmt.Errorf("Send tx failed: %v", errs)
}

func postRawTx(endpoint string, rawtx string) (string, error) {
	client := &http.Client{}
	var data = strings.NewReader(`{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["` + rawtx + `"],"id":1}`)
	req, err := http.NewRequest("POST", endpoint, data)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result interface{}
	err = json.Unmarshal(bodyText, result)
	if err != nil {
		return "", err
	}
	m, ok := result.(map[string]interface{})
	if ok {
		txhash, ok := m["result"].(string)
		if ok {
			return txhash, nil
		}
	}
	return "", fmt.Errorf("Decode response error")
}
