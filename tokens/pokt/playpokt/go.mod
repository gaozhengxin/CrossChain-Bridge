module github.com/gaozhengxin/CrossChain-Bridge/tokens/pokt/playpokt

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/pokt-network/pocket-core v0.0.0-20210521222253-8fba62eef22f
	github.com/tendermint/tendermint v0.33.7
)

replace github.com/tendermint/tendermint => github.com/pokt-network/tendermint v0.32.11-0.20210427155510-04e1c67f3eed // indirect
