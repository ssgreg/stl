package stl

import "github.com/gofrs/uuid"

// MergeTx prepares a single transaction from several simultaneous ones.
func MergeTx(ttx ...Tx) Tx {
	txMap := map[string]bool{}
	for _, tx := range ttx {
		for _, s := range tx.ListShared() {
			txMap[s] = false
		}
	}
	for _, tx := range ttx {
		for _, s := range tx.ListExclusive() {
			txMap[s] = true
		}
	}
	exs := make([]string, 0, len(txMap))
	shs := make([]string, 0, len(txMap))
	for k := range txMap {
		if txMap[k] {
			exs = append(exs, k)
		} else {
			shs = append(shs, k)
		}
	}

	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return &builder{
		id:  id.String(),
		shs: shs,
		exs: exs,
		v:   nil,
	}
}
