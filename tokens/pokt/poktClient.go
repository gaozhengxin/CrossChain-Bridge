package pokt

import (
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// GetLatestBlockNumberOf impl
// returns latest block number from a single endpoint
func (b *Bridge) GetLatestBlockNumberOf(apiAddress string) (uint64, error) {
	return 0, nil
}

// GetLatestBlockNumber impl
 // returns latest block number from one of the accesiable endpoints
func (b *Bridge) GetLatestBlockNumber() (uint64, error) {
	return 0, nil
}

// GetBlockHash takes a block number, returns block hash
func (b *Bridge) GetBlockHash(blockNumber uint64) string {
	return ""
}

// GetBlockTxids takes a block hash, returns block transaction ids in the block
func (b *Bridge) GetBlockTxids(blockHash string) ([]string, error) {
	return nil, nil
}

/*
GetBlockHash and GetBlockTxids is not mandantory for the bridge
They are used in scanning transactions

Alternatively, perhaps we can have functions like:
GetBlockTransactions(blockHash string) (txs []interface{}, error)
GetTransactionsRange(start, end uint64) (txs []interface{}, error)
GetTransactionsToRange(receiptAddress string, start, end uint64) (txs []interface{}, error)
*/

// GetTransaction returns transaction struct
func (b *Bridge) GetTransaction(txHash string) (interface{}, error) {
	return nil, nil
}

// GetTransactionStatus returns transaction status\
func (b *Bridge) GetTransactionStatus(txHash string) *tokens.TxStatus {
	// Either TxStatus.PrioriFinalized or TxStatus Confirmations is needed
	// depending on blockhchain protocol
	// if we can assert tx is finalized, fill PrioriFinalized with `true`
	// if we can't say tx is absolutely finalized (like in ethereum), fill Confirmations
	return &tokens.TxStatus {
		Receipt: nil,
		PrioriFinalized: false,
		Confirmations: 0,
		BlockHeight: 0,
		BlockHash: "",
		BlockTime: 0,
	}
}

// GetAddressBalance rerturns account POKT balance
func (b *Bridge) GetBalance(accountAddress string) (*big.Int, error) {
	return big.NewInt(0), nil
}

// GetTokenBalance not used in pokt bridge
func (b *Bridge) GetTokenBalance(tokenType, tokenAddress, accountAddress string) (*big.Int, error) {
	// No implement
	return big.NewInt(0), nil
}

// GetTokenSupply not used in pokt bridge
func (b *Bridge) GetTokenSupply(tokenType, tokenAddress string) (*big.Int, error) {
	// No implement
	return big.NewInt(0), nil
}

// SendTransaction sends a signed tx to fullnode to broadcast, returns txhash or an error
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	// TODO assert type of signedTx
	return "", nil
}