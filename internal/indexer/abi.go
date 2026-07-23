package indexer

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed abi/vault.abi.json
var vaultABIJSON []byte

func loadVaultABI() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(string(vaultABIJSON)))
}
