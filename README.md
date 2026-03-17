# flow

A Go package for building concurrent data pipelines.
You provide plain functions — the package handles all the channel wiring, goroutine management, and synchronization.

## Install

```sh
go get github.com/mycreepy/flow
```

## Usage

### Task

A `Task` is the core building block — a function that reads from an input channel and writes to an output channel.
Tasks must close `out` when done.

```go
double := flow.Task[int, int](func(in <-chan int, out chan<- int) {
    for v := range in {
        out <- v * 2
    }
    close(out)
})
```

### Concurrent

Fan out a task across multiple goroutines:

```go
heavy := flow.Concurrent(flow.Task[int, int](func(in <-chan int, out chan<- int) {
    for v := range in {
        time.Sleep(time.Second)
        out <- v * 2
    }
    close(out)
}), 4)

results := flow.FromValues(heavy, 1, 2, 3, 4) // processed by 4 workers concurrently
```

### Chaining same types

When all tasks share the same type, use `Chain`:

```go
pipeline := flow.Chain(
    flow.Task[int, int](func(in <-chan int, out chan<- int) {
        for v := range in { out <- v + 1 }
        close(out)
    }),
    flow.Task[int, int](func(in <-chan int, out chan<- int) {
        for v := range in { out <- v * 2 }
        close(out)
    }),
)

results := flow.FromValues(pipeline, 10) // [22]
```

### Chaining different types

Use `Pipe` to connect tasks with different input/output types:

```go
double := func(in <-chan int, out chan<- int) {
    for v := range in { out <- v * 2 }
    close(out)
}))

toString := func(in <-chan int, out chan<- string) {
    for v := range in { out <- fmt.Sprintf("%d", v) }
    close(out)
}))

pipeline := flow.Pipe(double, toString)
results := flow.FromValues(pipeline, 1, 2, 3) // ["2", "4", "6"]
```

For longer chains, nest `Pipe` calls:

```go
pipeline := flow.Pipe(flow.Pipe(parse, transform), encode)
```

### Custom producer

Use `FromChannel` when inputs come from a channel instead of a slice:

```go
double := func(in <-chan int, out chan<- int) {
    for v := range in { out <- v * 2 }
    close(out)
}))

ch := make(chan int)
go func() {
    defer close(ch)
    for i := range 5 {
        ch <- i
    }
}()

results := flow.FromChannel(double, ch) // [0, 2, 4, 6, 8]
```
