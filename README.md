# Software Transactional Locks

![Alt text](https://user-images.githubusercontent.com/1574981/41672584-ef9140a4-74c2-11e8-9da2-9ca926f77a53.png "(c) Ashley McNamara")
[![GoDoc](https://godoc.org/github.com/ssgreg/stl?status.svg)](https://godoc.org/github.com/ssgreg/stl)
[![Build Status](https://travis-ci.org/ssgreg/stl.svg?branch=master)](https://travis-ci.org/ssgreg/stl)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/stl)](https://goreportcard.com/report/github.com/ssgreg/stl)
[![Coverage Status](https://coveralls.io/repos/github/ssgreg/stl/badge.svg?branch=master)](https://coveralls.io/github/ssgreg/stl?branch=master)

Package `stl` provides multiple atomic dynamic shared/exclusive locks, based of [Software Transactional Memory](https://en.wikipedia.org/wiki/Software_transactional_memory) (STM) concurrency control mechanism.
In addition `stl` locks can take `context.Context` that allows to cancel or set a deadline for such locks.

Locks are fast and lightweight. The implementation requires only one `mutex` and one `channel` per vault (a set of locks with any number of resources).

## Installation

Install the package with:

```shell
go get github.com/ssgreg/stl
```

## TL;DR

`stl` can lock any number of resources atomically without a deadlock. Each resource can be locked in `exclusive` manner (only one locker can lock such resource at the same time) or in `shared` manner (all 'shared' lockers can lock such resource at the same time, a locker that wants to lock such resource 'exclusively' will wait them to finish).

You can also combine `shared` and `exclusive` resources while building a transaction or transactional locker:

```go
// A vault that holds all locked resources.
v := stl.NewVault()

// ...

locker := stl.New().Exclusive("terminal").Shared("network").ToLocker(v)
locker.Lock()
defer locker.Unlock()
```

It's also possible to call `locker.LockWithContext(ctx)` if you want to be able to cancel or set a deadline for locking operation. This will add additional flexibility to your applications.

## Example: Dining Philosophers Problem

In computer science, the [dining philosophers problem](https://en.wikipedia.org/wiki/Dining_philosophers_problem) is an example problem often used in concurrent algorithm design to illustrate synchronization issues and techniques for resolving them. Notably, with `stl`, the task can be solved elegantly comparing to any other “bare” solutions. The following is a short description of the problem:

Five philosophers are sitting at the circle dining table. There are five plates with spaghetti on it, and also five forks arranged so that every philosopher can take two nearest forks in both his hands. Philosophers either can think about their philosophical problems, or they can eat spaghetti. The thinkers are obligated to use two forks for eating. If a philosopher holds just one fork, he can’t eat or think. It's needed to organize their existence so that every philosopher could eat and think by turn, forever.

The task doesn’t sound that difficult, right? But aware of some pitfalls unseen from the first sight. The root of the problem are forks which can be considered as shared mutable resources for goroutines (philosophers). Any two neighbor thinkers are competing for the fork between them, and this enables such silly situations like “every philosopher has took the right fork, and they all stuck because no one could take the left fork anymore”. It's a deadlock. Thread starvation problem also can occur in a wrongly developed code, and, ironically, it will result in “starvation” of some philosophers: while part of them eat and think normally, other part can acquire resources hardly ever. So in good solutions all philosophers should pass their think-eat iterations almost equally.

Let’s see how we can do it with `stl`.

Firstly we need to represent forks (resources) in our concurrent model. To distinguish different forks, we’ll assign some label to each. There is no need to create and keep resources itself using `stl`. Labels (or names) are enoughs.

```go
// Two forks per each five philosophers.
resources := [][]string{
    {"fork_1", "fork_2"},
    {"fork_2", "fork_3"},
    {"fork_3", "fork_4"},
    {"fork_4", "fork_5"},
    {"fork_5", "fork_1"},
}
```

Let's continue with constructing a transaction (or a transaction locker like in this case). When the philosopher changes his activity from `thinking` to `eating`, he tries to take the left and the right fork `exclusively`. If he successfully takes (locks) both, he spends some time eating his spaghetti. If any of the forks is taken by a neighbor, our philosopher should wait for any other philosopher to finish eating (unlock).

This is how transaction lockers for each philosophers can be created.

```go
// A vault that holds all locked resources.
v := stl.NewVault()

// ...

for n := 0; n < 5; n++ {
    // ...
    // A locker that can exclusively lock/unlock both forks atomically.
    locker := stl.New().Exclusive(resources[n][0]).Exclusive(resources[n][1]).ToLocker(v)
    // ...
}
```

To change philosopher's activity (exclusively lock both specified resources) we need to call `locker.Lock()` and `locker.Unlock()` in the end. It's also possible to call `locker.LockWithContext(ctx)` if we want to be able to cancel or set a deadline for locking operation because it could take some time waiting for other actors to unlock used resources.

```go
// Philosopher is thinking here...
// ...

// Now he decided to take forks and eat a little bit of his spaghetti.
locker.Lock()
defer locker.Unlock()

// Philosopher is eating here...
// ...
```

But now we have to stop and understand what this transaction locker will do. Suppose, both forks were free, then the locker will successfully lock both resources their represent. However it’s more likely one of the forks was already taken by someone else. _“All or nothing”_, - this principle works well with STM mechanism. If any fork of the two was already taken, the transaction will be restarted until the locker will successfully lock both resources. We can consider the transaction will be successful if and only if all locker's resources will be successfully locked. In other words the transaction locker won’t proceed further if some of the forks was in the undesired state.

We are about to finish our solution. Think-eat function:

```go
do := func(locker Locker) {
    // Think for 300 ms.
    time.Sleep(time.Millisecond * 300)

    // Wait for free forks.
    locker.Lock()
    defer locker.Unlock()

    // Eat for 100ms.
    time.Sleep(time.Millisecond * 100)
}
```

Evaluation:

```go
for n := 0; n < 5; n++ {
    // Each of five philosopher live in it's own goroutine.
    go func(n int) {
        // A locker that can exclusively lock/unlock both forks atomically.
        locker := New().Exclusive(resources[n][0]).Exclusive(resources[n][1]).ToLocker(v)

        // Think-Eat for five times.
        for i := 0; i < 5; i++ {
            do(locker)
        }
    }(n)
}
```

If you run the code, you’ll see the following output (with my comments):

```
19.663µs        | Fifth  is thinking 1th time
57.726µs        | Second is thinking 1th time
131.391µs       | First  is thinking 1th time
98.969µs        | Third  is thinking 1th time
206.433µs       | Fourth is thinking 1th time
```

> All five philosophers are starting to think.

```
304.330265ms    | Fourth is eating 1th time, was starving for 102.93µs
304.304333ms    | First  is eating 1th time, was starving for 90.307µs
```

> 300ms later. Fourth and First are eating now. They took forks with numbers 1, 2, 4, 5. Second and Third can't eat with one fork 3. Fifth can't even take a single fork.

```
404.955441ms    | Third  is eating 1th time, was starving for 100.733725ms
404.952165ms    | Fourth is thinking 2th time
404.986976ms    | First  is thinking 2th time
405.005764ms    | Fifth  is eating 1th time, was starving for 100.758988ms
```

> 100ms later. Fourth and First finished eating. They put their forks. Forks with number 3, 4, 5, 1 was taken by Third and Fifth. Second is really unlucky guy.

```
505.144439ms    | Third  is thinking 2th time
505.153147ms    | Fifth  is thinking 2th time
505.195732ms    | Second is eating 1th time, was starving for 200.986826ms
```

> 100ms later. Third and Fifth finished their meal. Second is the only dining philosopher now. He was starving for 200ms while waiting for others.

```
608.972697ms    | Second is thinking 2th time
705.095973ms    | Fourth is eating 2th time, was starving for 8.551µs
705.104647ms    | First  is eating 2th time, was starving for 12.506µs
808.134488ms    | First  is thinking 3th time
808.187194ms    | Fourth is thinking 3th time
808.194107ms    | Third  is eating 2th time, was starving for 110.638µs
808.228277ms    | Fifth  is eating 2th time, was starving for 131.242µs
913.369796ms    | Fifth  is thinking 3th time
913.417331ms    | Third  is thinking 3th time
913.433097ms    | Second is eating 2th time, was starving for 73.613µs
1.016172052s    | Second is thinking 3th time
1.113263947s    | Fourth is eating 3th time, was starving for 6.818µs
1.113268306s    | First  is eating 3th time, was starving for 10.244µs
1.216127462s    | First  is thinking 4th time
1.216138289s    | Fourth is thinking 4th time
1.216149318s    | Third  is eating 3th time, was starving for 23.142µs
1.216168988s    | Fifth  is eating 3th time, was starving for 57.036µs
1.31888003s     | Fifth  is thinking 4th time
1.318897601s    | Third  is thinking 4th time
1.318909899s    | Second is eating 3th time, was starving for 87.83µs
1.421858869s    | Second is thinking 4th time
1.518619148s    | Fourth is eating 4th time, was starving for 8.182µs
1.518632238s    | First  is eating 4th time, was starving for 6.313µs
1.622124278s    | Fourth is thinking 5th time
1.622126414s    | First  is thinking 5th time
1.622132975s    | Fifth  is eating 4th time, was starving for 15.59µs
1.62213733s     | Third  is eating 4th time, was starving for 27.384µs
1.725392718s    | Fifth  is thinking 5th time
1.725411187s    | Third  is thinking 5th time
1.725421148s    | Second is eating 4th time, was starving for 17.534µs
1.829165491s    | Second is thinking 5th time
1.926475691s    | Fourth is eating 5th time, was starving for 10.229µs
1.926481498s    | First  is eating 5th time, was starving for 6.972µs
2.028900949s    | Third  is eating 5th time, was starving for 39.312µs
2.028929715s    | Fifth  is eating 5th time, was starving for 64.779µs
2.13402499s     | Second is eating 5th time, was starving for 6.024µs
```

> No philosophers were waiting for others. Their found they think-eat balance.

## Conclusion

`stl` provides you some handy way to build reliable deadlock-free applications.

## Additional materials

* [Software Transactional Memory (Wikipedia)](https://en.wikipedia.org/wiki/Software_transactional_memory)
* [Dining Philosophers Problem (Wikipedia)](https://en.wikipedia.org/wiki/Dining_philosophers_problem)
* [TS 19841:2015 Transactional Memory](https://www.iso.org/standard/66343.html)
* [Beyound locks: Software Transactional Memory](https://bartoszmilewski.com/2010/09/11/beyond-locks-software-transactional-memory/)
