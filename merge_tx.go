package stl

// MergeTx prepares a single transaction from several simultaneous ones.
func MergeTx(ttx ...Tx) Tx {
	txMap := map[string]bool{}
	for _, tx := range ttx {
		for _, s := range tx.ListShared() {
			// Ensure exclusive key was not set before.
			if !txMap[s] {
				txMap[s] = false
			}
		}
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

	return &builder{
		shs: shs,
		exs: exs,
		v:   nil,
	}
}