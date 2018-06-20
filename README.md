# Software Transactional Locks

[![GoDoc](https://godoc.org/github.com/ssgreg/stl?status.svg)](https://godoc.org/github.com/ssgreg/stl)
[![Build Status](https://travis-ci.org/ssgreg/stl.svg?branch=master)](https://travis-ci.org/ssgreg/stl)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/stl)](https://goreportcard.com/report/github.com/ssgreg/stl)
[![Coverage Status](https://coveralls.io/repos/github/ssgreg/stl/badge.svg?branch=master)](https://coveralls.io/github/ssgreg/stl?branch=master)

Package `stl` provides multiple atomic dynamic shared/exclusive locks, based of [Software Transactional Memory](https://en.wikipedia.org/wiki/Software_transactional_memory) (STM) concurrency control mechanism.
In addition `stl` locks take `context.Context` that allows you to cancel or set a deadline to such locks.

Locks are fast and lightweight. The implementation requires only one `mutex` and one `channel` per vault (a set of locks with any number of resources).

## Installation

Install the package with:

```shell
go get github.com/ssgreg/stl
```

## Dining Philosophers problem