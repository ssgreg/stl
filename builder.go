package stl

import "context"

// New creates an instance of default Builder.
func New() Builder {
	return &builder{}
}

type builder struct {
	shs []string
	exs []string
	v   Vault
}

func (t *builder) Shared(name string) Builder {
	t.shs = append(t.shs, name)

	return t
}

func (t *builder) Exclusive(name string) Builder {
	t.exs = append(t.exs, name)

	return t
}

func (t *builder) ListShared() []string {
	return t.shs
}

func (t *builder) ListExclusive() []string {
	return t.exs
}

func (t *builder) ToTx() Tx {
	return t
}

func (t *builder) ToLocker(v Vault) Locker {
	t.v = v
	return t
}

func (t *builder) Lock() {
	_ = t.v.Lock(context.Background(), t)
}

func (t *builder) Unlock() {
	t.v.Unlock(t)
}

func (t *builder) LockWithContext(ctx context.Context) error {
	return t.v.Lock(ctx, t)
}
