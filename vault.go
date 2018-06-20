package stl

import (
	"context"
	"sync"
)

// NewVault creates an instance of Vault.
func NewVault() Vault {
	return &vault{
		wait: make(chan struct{}, 0),
		rs:   make(map[string]*resource),
	}
}

type vault struct {
	mu   sync.Mutex
	wait chan struct{}
	rs   map[string]*resource
}

// Lock locks the resources according to specified Tx.
func (v *vault) Lock(ctx context.Context, tx Tx) error {
	for {
		if wait := v.tryLock(tx); wait != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-wait:
				continue
			}
		}

		return nil
	}
}

// Unlock unlocks the resources according to specified Tx.
func (v *vault) Unlock(tx Tx) {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, rn := range tx.ListExclusive() {
		delete(v.rs, rn)
	}

	for _, rn := range tx.ListShared() {
		r := v.rs[rn]
		if r.readers == 1 {
			delete(v.rs, rn)
		} else {
			r.readers--
		}
	}

	v.notifyWait()
}

func (v *vault) tryLock(tx Tx) <-chan struct{} {
	v.mu.Lock()
	defer v.mu.Unlock()

	for _, rn := range tx.ListExclusive() {
		if _, ok := v.rs[rn]; ok {
			return v.exposeWait()
		}
	}

	for _, rn := range tx.ListShared() {
		if r, ok := v.rs[rn]; ok {
			if r.writer {
				return v.exposeWait()
			}
		}
	}

	for _, rn := range tx.ListExclusive() {
		v.rs[rn] = &resource{0, true}
	}

	for _, rn := range tx.ListShared() {
		if r, ok := v.rs[rn]; ok {
			r.readers++
		} else {
			v.rs[rn] = &resource{1, false}
		}
	}

	return nil
}

func (v *vault) exposeWait() <-chan struct{} {
	if v.wait == nil {
		v.wait = make(chan struct{}, 0)
	}

	return v.wait
}

func (v *vault) notifyWait() {
	if v.wait != nil {
		close(v.wait)
		v.wait = nil
	}
}

type resource struct {
	readers int
	writer  bool
}
