/*
Package `stl` provides multiple atomic dynamic shared/exclusive locks, based of Software Transactional Memory (STM) concurrency control mechanism. In addition `stl` locks can take `context.Context` that allows to cancel or set a deadline for such locks.

Locks are fast and lightweight. The implementation requires only one `mutex` and one `channel` per vault (a set of locks with any number of resources).

`stl` can lock any number of resources atomically without a deadlock. Each resource can be locked in `exclusive` manner (only one locker can lock such resource at the same time) or in `shared` manner (all 'shared' lockers can lock such resource at the same time, a locker that wants to lock such resource 'exclusively' will wait them to finish).

You can also combine `shared` and `exclusive` resources while building a transaction or transactional locker:

	// A vault that holds all locked resources.
	v := stl.NewVault()

	// ...

	locker := stl.New().Exclusive("terminal").Shared("network").ToLocker(v)
	locker.Lock()
	defer locker.Unlock()

It's also possible to call `locker.LockWithContext(ctx)` if you want to be able to cancel or set a deadline for locking operation. This will add additional flexibility to your applications.
*/
package stl
