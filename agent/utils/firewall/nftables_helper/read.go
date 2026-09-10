package nftables_helper

import "errors"

var ErrChainNotFound = errors.New("nftables chain is not initialized")

func ReadChain(run func(...string) (string, error), family, table, chain string) (string, error) {
	output, exists, err := readNftObject(run, "-a", "list", "chain", family, table, chain)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrChainNotFound
	}
	return output, nil
}
