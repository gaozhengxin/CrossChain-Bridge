package pokt

import (
	"fmt"

	"github.com/anyswap/CrossChain-Bridge/common"
)

var TypePOKT2ETHAddressBinding = "pokt-eth-bindaddress"

// Address binding agreement is registered at CrossChain-Bridge server
// and saved in server's database
// will be managed by a smart contract in future
type AddressBindingAgreement struct {
	ETHAddress common.Address
	// TODO define type of pokt address
	PoktAddress interface{}
}

func (agreement AddressBindingAgreement) Type() string {
	return TypePOKT2ETHAddressBinding
}

func (agreement AddressBindingAgreement) Key() string {
	// TODO pokt address type would be better to implement String() string
	return fmt.Sprintf("%s", agreement.PoktAddress)
}

func (agreement AddressBindingAgreement) Value() interface{} {
	return agreement.ETHAddress
}
