package stl

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ssgreg/bottleneck"
	"github.com/stretchr/testify/require"
)

func TestShared(t *testing.T) {
	const Dur = time.Millisecond * 100
	const Workers = 10

	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(Workers)
	bc.TimeSlice(bottleneck.Index0)
	var callCounter uint32

	for i := 0; i < Workers; i++ {
		go func(i int) {
			defer wg.Done()
			locker := New().Shared("shared").ToLocker(v)
			locker.Lock()
			defer locker.Unlock()
			atomic.AddUint32(&callCounter, 1)
			time.Sleep(Dur)
		}(i)
	}

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, Workers, callCounter)
	require.True(t, stats[0].Duration < Dur+time.Millisecond*20 && stats[0].Duration > Dur, "should take ~100ms")
}

func TestExclusive(t *testing.T) {
	const Dur = time.Millisecond * 10
	const Workers = 10

	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(Workers)
	bc.TimeSlice(bottleneck.Index0)
	callCounter := 0

	for i := 0; i < Workers; i++ {
		go func(i int) {
			defer wg.Done()
			tr := New().Exclusive("exclusive").ToTx()
			err := v.Lock(context.Background(), tr)
			require.NoError(t, err)
			defer v.Unlock(tr)
			callCounter++
			time.Sleep(Dur)
		}(i)
	}

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, Workers, callCounter)
	require.True(t, stats[0].Duration < Dur*Workers+time.Millisecond*50 && stats[0].Duration > Dur*Workers)
}

func TestSharedWaitsExclusive(t *testing.T) {
	const Dur = time.Millisecond * 50
	const WaitDur = Dur / 2

	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(2)

	bc.TimeSlice(bottleneck.Index0)
	callCounter := 0

	go func() {
		defer wg.Done()
		locker := New().Exclusive("resource").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		callCounter++
		time.Sleep(Dur)
	}()

	time.Sleep(WaitDur)

	go func() {
		defer wg.Done()
		locker := New().Shared("resource").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		callCounter++
		time.Sleep(Dur)
	}()

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, 2, callCounter)
	require.True(t, stats[0].Duration < Dur*2+time.Millisecond*20 && stats[0].Duration > Dur*2)
}

func TestExclusiveWaitsShared(t *testing.T) {
	const Dur = time.Millisecond * 50
	const WaitDur = Dur / 2

	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(2)

	bc.TimeSlice(bottleneck.Index0)
	callCounter := 0

	go func() {
		defer wg.Done()
		locker := New().Shared("resource").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		callCounter++
		time.Sleep(Dur)
	}()

	time.Sleep(WaitDur)

	go func() {
		defer wg.Done()
		locker := New().Exclusive("resource").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		callCounter++
		time.Sleep(Dur)
	}()

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, 2, callCounter)
	require.True(t, stats[0].Duration < Dur*2+time.Millisecond*20 && stats[0].Duration > Dur*2)
}

func TestExclusiveAndSharedInParallelWithDiffExclusiveAndSameShared(t *testing.T) {
	const Dur = time.Millisecond * 100

	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(2)

	bc.TimeSlice(bottleneck.Index0)
	var callCounter uint32

	go func() {
		defer wg.Done()
		locker := New().Shared("same_shared").Exclusive("exclusive").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		atomic.AddUint32(&callCounter, 1)
		time.Sleep(Dur)
	}()

	go func() {
		defer wg.Done()
		locker := New().Shared("same_shared").Exclusive("diff_exclusive").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		atomic.AddUint32(&callCounter, 1)
		time.Sleep(Dur)
	}()

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, 2, callCounter)
	require.True(t, stats[0].Duration < Dur+time.Millisecond*20 && stats[0].Duration > Dur)
}

func TestStacked(t *testing.T) {
	const Dur = time.Millisecond * 100
	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(2)

	bc.TimeSlice(bottleneck.Index0)
	var callCounter uint32

	go func() {
		defer wg.Done()
		locker := NewStacked().Shared("a1").Exclusive("s1").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		atomic.AddUint32(&callCounter, 1)
		time.Sleep(Dur)
	}()

	go func() {
		defer wg.Done()
		locker := NewStacked().Shared("a2").Exclusive("s1").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		atomic.AddUint32(&callCounter, 1)
		time.Sleep(Dur)
	}()

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, 2, callCounter)
	require.True(t, stats[0].Duration < Dur+time.Millisecond*20 && stats[0].Duration > Dur)
}

func TestContextCancel(t *testing.T) {
	const Dur = time.Millisecond * 50
	const WaitDur = Dur / 4

	bc := bottleneck.NewCalculator()
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(2)

	ctx, cancel := context.WithCancel(context.Background())

	bc.TimeSlice(bottleneck.Index0)
	callCounter := 0

	go func() {
		defer wg.Done()
		locker := New().Exclusive("resource").ToLocker(v)
		locker.Lock()
		defer locker.Unlock()

		callCounter++
		time.Sleep(Dur)
	}()

	time.Sleep(WaitDur)

	go func() {
		defer wg.Done()
		locker := New().Exclusive("resource").ToLocker(v)
		err := locker.LockWithContext(ctx)
		require.EqualError(t, err, "context canceled")
		if err != nil {
			return
		}

		defer locker.Unlock()
		callCounter++
	}()

	time.Sleep(WaitDur)
	cancel()

	wg.Wait()

	stats := bc.Stats()
	fmt.Println("Duration: ", stats[0].Duration)
	require.EqualValues(t, 1, callCounter)
	require.True(t, stats[0].Duration < Dur+time.Millisecond*20 && stats[0].Duration > Dur)
}

func TestDiningPhilosophers(t *testing.T) {
	v := NewVault()

	var wg sync.WaitGroup
	wg.Add(5)

	start := time.Now()

	do := func(name string, i int, locker Locker) {
		// Think for 300 ms.
		fmt.Printf("%v \t| %s is thinking %dth time\n", time.Now().Sub(start), name, i+1)
		time.Sleep(time.Millisecond * 300)

		// Wait for lock (starvation duration).
		timeStartStaving := time.Now()
		locker.Lock()
		defer locker.Unlock()

		// Eat for 100ms.
		fmt.Printf("%v \t| %s is eating %dth time, was starving for %v\n", time.Now().Sub(start), name, i+1, time.Now().Sub(timeStartStaving))
		time.Sleep(time.Millisecond * 100)
	}

	names := []string{"First ", "Second", "Third ", "Fourth", "Fifth "}
	resources := [][]string{
		{"fork_1", "fork_2"},
		{"fork_2", "fork_3"},
		{"fork_3", "fork_4"},
		{"fork_4", "fork_5"},
		{"fork_5", "fork_1"},
	}

	for n := 0; n < 5; n++ {
		// Each of five philosopher live in it's own goroutine.
		go func(n int) {
			defer wg.Done()

			// A locker that can exclusively lock/unlock both forks atomically.
			locker := New().Exclusive(resources[n][0]).Exclusive(resources[n][1]).ToLocker(v)

			// Think/Eat for five times.
			for i := 0; i < 5; i++ {
				do(names[n], i, locker)
			}
		}(n)
	}

	wg.Wait()
}
