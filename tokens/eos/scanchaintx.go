package eos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"
)

var (
	quickSyncFinish  bool
	quickSyncWorkers = uint64(4)

	maxScanHeight          = uint64(100)
	retryIntervalInScanJob = 3 * time.Second
	restIntervalInScanJob  = 3 * time.Second

	getActionsOffset uint64 = 10
)

func (b *Bridge) getStartAndLatestHeight() (start, latest uint64) {
	startHeight := tools.GetLatestScanHeight(b.IsSrc)

	chainCfg := b.GetChainConfig()
	confirmations := *chainCfg.Confirmations
	initialHeight := *chainCfg.InitialHeight

	latest = tools.LoopGetLatestBlockNumber(b)

	switch {
	case startHeight != 0:
		start = startHeight
	case initialHeight != 0:
		start = initialHeight
	default:
		if latest > confirmations {
			start = latest - confirmations
		}
	}
	if start < initialHeight {
		start = initialHeight
	}
	if start+maxScanHeight < latest {
		start = latest - maxScanHeight
	}
	return start, latest
}

func (b *Bridge) getStartSequence() uint64 {
	token := b.GetTokenConfig(PairID)
	if token.InitialSeq == nil {
		return 0
	}
	return *token.InitialSeq
}

// StartChainTransactionScanJob scan job
func (b *Bridge) StartChainTransactionScanJob() {
	b.StartAccountActionScanJob()
}

// StartAccountActionScanJob scan actions from account
// Actions ars indexed by  sequence
func (b *Bridge) StartAccountActionScanJob() {
	chainName := b.ChainConfig.BlockChain
	log.Infof("[scanchain] start %v scan chain job", chainName)

	token := b.GetTokenConfig(PairID)
	depositAddress := token.DepositAddress

	startSeq := b.getStartSequence()

	for {
		resp, err := b.GetActions(depositAddress, int64(startSeq), int64(getActionsOffset))
		if err != nil {
			log.Error("get actions fail", "start", startSeq, "offset", getActionsOffset, "error", err)
			time.Sleep(retryIntervalInScanJob)
			continue
		}
		for _, action := range resp.Actions {
			if uint64(action.AccountSeq) > startSeq {
				startSeq = uint64(action.AccountSeq)
			}
			txhash := action.Trace.TransactionID.String()
			b.processTransaction(txhash)
		}
		startSeq++
		if len(resp.Actions) == 0 {
			time.Sleep(time.Second * 15)
		}
	}
}

// StartBlockTransactionScanJob scan block job
func (b *Bridge) StartBlockTransactionScanJob() {
	chainName := b.ChainConfig.BlockChain
	log.Infof("[scanchain] start %v scan chain job", chainName)

	cli := b.GetClient()
	ctx := context.Background()

	start, latest := b.getStartAndLatestHeight()
	_ = tools.UpdateLatestScanInfo(b.IsSrc, start)
	log.Infof("[scanchain] start %v scan chain loop from %v latest=%v", chainName, start, latest)

	errorSubject := fmt.Sprintf("[scanchain] get %v block failed", chainName)
	scanSubject := fmt.Sprintf("[scanchain] scanned %v block", chainName)

	if latest > start {
		go b.quickSync(context.Background(), nil, start, latest+1)
	} else {
		quickSyncFinish = true
	}

	stable := latest

	scannedBlocks := tools.NewCachedScannedBlocks(67)
	var quickSyncCtx context.Context
	var quickSyncCancel context.CancelFunc
	for {
		latest = tools.LoopGetLatestBlockNumber(b)
		if stable+maxScanHeight < latest {
			if quickSyncCancel != nil {
				select {
				case <-quickSyncCtx.Done():
				default:
					log.Warn("cancel quick sync range", "stable", stable, "latest", latest)
					quickSyncCancel()
				}
			}
			quickSyncCtx, quickSyncCancel = context.WithCancel(context.Background())
			go b.quickSync(quickSyncCtx, quickSyncCancel, stable+1, latest)
			stable = latest
		}
		for h := stable; h <= latest; {
			// get block by height
			blockResp, err := cli.GetBlockByNum(ctx, uint32(h))
			if err != nil {
				log.Error(errorSubject, "height", h, "err", err)
				time.Sleep(retryIntervalInScanJob)
				continue
			}
			blockID := blockResp.ID.String()
			if scannedBlocks.IsBlockScanned(blockID) {
				h++
				continue
			}
			// get transactions from all blocks
			for _, tx := range blockResp.Transactions {
				b.processTransaction(tx.Transaction.ID.String())
			}
			scannedBlocks.CacheScannedBlock(blockID, h)

			scannedBlocks.CacheScannedBlock(blockID, h)
			log.Info(scanSubject, "blockId", blockID, "height", h, "transactions", len(blockResp.Transactions))
			h++
			time.Sleep(60 * time.Second)
		}
		stable = latest
		if quickSyncFinish {
			_ = tools.UpdateLatestScanInfo(b.IsSrc, stable)
		}
		time.Sleep(restIntervalInScanJob)
	}
}

func (b *Bridge) quickSync(ctx context.Context, cancel context.CancelFunc, start, end uint64) {
	chainName := b.ChainConfig.BlockChain
	log.Printf("[scanchain] begin %v syncRange job. start=%v end=%v", chainName, start, end)
	count := end - start
	workers := quickSyncWorkers
	if count < 10 {
		workers = 1
	}
	step := count / workers
	wg := new(sync.WaitGroup)
	wg.Add(int(workers))
	for i := uint64(0); i < workers; i++ {
		wstt := start + i*step
		wend := start + (i+1)*step
		if i+1 == workers {
			wend = end
		}
		go b.quickSyncRange(ctx, i+1, wstt, wend, wg)
	}
	wg.Wait()
	if cancel != nil {
		cancel()
	} else {
		quickSyncFinish = true
	}
	log.Printf("[scanchain] finish %v syncRange job. start=%v end=%v", chainName, start, end)
}

func (b *Bridge) quickSyncRange(ctx context.Context, idx, start, end uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	chainName := b.ChainConfig.BlockChain
	log.Printf("[scanchain] id=%v begin %v syncRange start=%v end=%v", idx, chainName, start, end)

	errorSubject := fmt.Sprintf("[scanchain] get %v block failed", chainName)

	for h := start; h < end; {
		select {
		case <-ctx.Done():
			break
		default:
		}

		// get block by height
		blockResp, err := cli.GetBlockByNum(ctx, uint32(h))
		if err != nil {
			log.Error(errorSubject, "height", h, "err", err)
			time.Sleep(retryIntervalInScanJob)
			continue
		}
		blockID := blockResp.ID.String()
		// get transactions from all blocks
		for _, tx := range blockResp.Transactions {
			b.processTransaction(tx.Transaction.ID.String())
		}

		log.Tracef("[scanchain] id=%v scanned %v block, height=%v hash=%v txs=%v", idx, chainName, h, blockID, len(blockResp.Transactions))
		h++
	}

	log.Printf("[scanchain] id=%v finish %v syncRange start=%v end=%v", idx, chainName, start, end)
}
