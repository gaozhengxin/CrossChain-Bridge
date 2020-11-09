package filecoin

import (
	"context"
	"fmt"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"

	filApi "github.com/filecoin-project/lotus/api"
)

var (
	scannedTxs = tools.NewCachedScannedTxs(300)
)

// StartPoolTransactionScanJob scan job
func (b *Bridge) StartPoolTransactionScanJob() {
	chainName := b.ChainConfig.BlockChain
	log.Infof("[scanpool] start scan %v tx pool job", chainName)
	errorSubject := fmt.Sprintf("[scanpool] get %v pool txs error", chainName)
	scanSubject := fmt.Sprintf("[scanpool] scanned %v tx", chainName)
	err := b.StartListenMpool(context.Background(), func(mup filApi.MpoolUpdate) {
		txid := mup.Message.Cid().String()
		if scannedTxs.IsTxScanned(txid) {
			return
		}
		log.Trace(scanSubject, "txid", txid)
		b.processTransaction(txid)
		scannedTxs.CacheScannedTx(txid)
	})
	if err != nil {
		log.Error(errorSubject, "err", err)
	}
}
