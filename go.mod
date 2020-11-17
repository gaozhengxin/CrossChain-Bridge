module github.com/anyswap/CrossChain-Bridge

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/blocknetdx/btcd v0.20.1-beta.0.20200618032145-59a950423708
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/btcsuite/btcwallet/wallet/txauthor v1.0.0
	github.com/btcsuite/btcwallet/wallet/txrules v1.0.0
	github.com/btcsuite/btcwallet/wallet/txsizes v1.0.0
	github.com/eoscanada/eos-go v0.9.0
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/filecoin-project/filecoin-ffi v0.30.4-0.20200910194244-f640612a1a1f // indirect
	github.com/filecoin-project/go-address v0.0.4
	github.com/filecoin-project/go-jsonrpc v0.1.2-0.20201008195726-68c6a2704e49
	github.com/filecoin-project/go-state-types v0.0.0-20201013222834-41ea465f274f
	github.com/filecoin-project/lotus v1.1.2
	github.com/fsn-dev/fsn-go-sdk v0.0.0-20200924054943-42f8aa0d3973
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/rpc v1.2.0
	github.com/ipfs/go-cid v0.0.7
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jonboulle/clockwork v0.2.0 // indirect
	github.com/jordan-wright/email v0.0.0-20200602115436-fd8a7622303e
	github.com/lestrrat-go/file-rotatelogs v0.0.0-20201029035330-b789b39afbd7
	github.com/lestrrat-go/strftime v1.0.1 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/pborman/uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/crypto v0.0.0-20201112155050-0c6587e931a9
	golang.org/x/sys v0.0.0-20201113233024-12cec1faf1ba // indirect
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
)

replace github.com/anyswap/CrossChain-Bridge => ./
