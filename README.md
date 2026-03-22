# flow

[![Go Reference](https://pkg.go.dev/badge/github.com/mycreepy/flow.svg)](https://pkg.go.dev/github.com/mycreepy/flow)
[![Go Report Card](https://goreportcard.com/badge/github.com/mycreepy/flow?style=flat-square)](https://goreportcard.com/report/github.com/mycreepy/flow)
[![Go Build & Test](https://github.com/mycrEEpy/flow/actions/workflows/build.yml/badge.svg)](https://github.com/mycrEEpy/flow/actions/workflows/build.yml)
[![Go Coverage](https://github.com/mycreepy/flow/wiki/coverage.svg)](https://raw.githack.com/wiki/mycreepy/flow/coverage.html)

[Gopher with a pipeline](https://i.ibb.co/YMXsPL6/go-flow-2-cropped.png)

A Go package for building concurrent data pipelines.
You provide plain functions — the package handles all the channel wiring, goroutine management, and synchronization.

## Install

```sh
go get github.com/mycreepy/flow
```

## Usage

### Task

A `Task` is the core building block — a function that reads from an input channel and writes to an output channel.
Tasks must close the `out` channel when done, including when returning an error.
Returning a non-nil error signals that processing has failed; closing the `out` channel ensures downstream tasks and goroutines terminate cleanly.

```go
double := flow.Task[int, int](func(in <-chan int, out chan<- int) error {
    defer close(out)
	for v := range in {
        out <- v * 2
    }
    return nil
})
```

The generic type definition of `Task` can be inferred most of the time as seen in the next examples. 

### Concurrent

Fan out a task across multiple goroutines:

```go
heavy := flow.Concurrent(4, func(in <-chan int, out chan<- int) error {
    defer close(out)
	for v := range in {
        time.Sleep(time.Second)
        out <- v * 2
    }
    return nil
})

results, err := flow.FromValues(heavy, 1, 2, 3, 4) // processed by 4 workers concurrently
```

### Chain & Pipe

When all tasks share the same type, use `Chain`:

```go
pipeline := flow.Chain(
    func(in <-chan int, out chan<- int) error {
        defer close(out)
        for v := range in { out <- v + 1 }
        return nil
    },
    func(in <-chan int, out chan<- int) error {
        defer close(out)
		for v := range in { out <- v * 2 }
        return nil
    },
)

results, err := flow.FromValues(pipeline, 10) // [22]
```

Use `Pipe` to connect tasks with different input/output types:

```go
double := func(in <-chan int, out chan<- int) error {
    defer close(out)
	for v := range in { out <- v * 2 }
    return nil
}

toString := func(in <-chan int, out chan<- string) error {
    defer close(out)
	for v := range in { out <- fmt.Sprintf("%d", v) }
    return nil
}

pipeline := flow.Pipe(double, toString)
results, err := flow.FromValues(pipeline, 1, 2, 3) // ["2", "4", "6"]
```

For longer chains, nest `Pipe` calls:

```go
pipeline := flow.Pipe(flow.Pipe(parse, transform), encode)
```

### Helper functions

Filter items in a pipeline with `Filter`:

```go
pipeline := flow.Filter(func(v int) (bool, error) { return v%2 == 0, nil })

results, err := flow.FromValues(pipeline, 1, 2, 3, 4, 5, 6) // [2, 4, 6]
```

Apply a function to each item with `ForEach`:

```go
pipeline := flow.ForEach(func(v int) (string, error) { return fmt.Sprintf("%d", v), nil })

results, err := flow.FromValues(pipeline, 1, 2, 3) // ["1", "2", "3"]
```

Duplicate items to a slice with `Append`:

```go
var collected []int
pipeline := flow.Append(&collected)

results, err := flow.FromValues(pipeline, 1, 2, 3) // [1, 2, 3]
// collected == [1, 2, 3]
```

Duplicate items to a side channel with `Tee`:

```go
tee := make(chan int, 100)
pipeline := flow.Tee(tee)

go func() {
	for v := range tee {
		fmt.Println(v) // 1, 2, 3
    }
}
}()

results, err := flow.FromValues(pipeline, 1, 2, 3) // [1, 2, 3]
close(tee)
```

Log progress every N items with `LogEveryN`:

```go
pipeline := flow.LogEveryN[int](1000, slog.Default(), "processed items")

results, err := flow.FromSlice(pipeline, veryBigSlice)
```

### Error handling

When a task returns an error, it must still close `out` (use `defer close(out)`).
This ensures all downstream goroutines drain and terminate — no goroutine leaks.
The error is propagated through `Pipe`, `Chain`, `Concurrent`, and returned by `FromValues`, `FromSlice` & `FromChannel`.

```go
pipeline := flow.Pipe(
    func(in <-chan int, out chan<- int) error {
        defer close(out)
        for v := range in {
            if v < 0 {
                return errors.New("negative value")
            }
            out <- v
        }
        return nil
    },
    func(in <-chan int, out chan<- string) error {
        defer close(out)
        for v := range in { out <- fmt.Sprintf("%d", v) }
        return nil
    },
)

results, err := flow.FromValues(pipeline, 1, -1, 3)
if err != nil {
    slog.Error("data pipeline failed", "error", err) // "negative value"
}
```
