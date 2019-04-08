package stl

// Join prepares a single transaction from several simultaneous ones.
func Join(ttx ...Tx) Tx {
	shMap := map[string]bool{}
	exMap := map[string]bool{}
	for _, tx := range ttx {
		for _, s := range tx.ListShared() {
			shMap[s] = true
		}
		for _, s := range tx.ListExclusive() {
			exMap[s] = true
		}
	}
	exs := make([]string, 0, len(exMap))
	for k := range exMap {
		exs = append(exs, k)
		if shMap[k] {
			shMap[k] = false
		}
	}
	shs := make([]string, 0, len(shMap))
	for k := range shMap {
		if shMap[k] {
			shs = append(shs, k)
		}
	}

	return &builder{
		shs: shs,
		exs: exs,
		v:   nil,
	}
}