# flow

A Go package for building concurrent data pipelines.
You provide plain functions — the package handles all the channel wiring, goroutine management, and synchronization.

## Install

```sh
go get github.com/mycreepy/flow
```

## Usage

### Basic pipeline

Provide a `func(In) Out` and the package does the rest:

```go
double := flow.Func(func(v int) int { return v * 2 })
results := flow.Run(double, 1, 2, 3) // [2, 4, 6]
```

### Chaining same types

When all tasks share the same type, use `Chain`:

```go
pipeline := flow.Chain(
    flow.Func(func(v int) int { return v + 1 }),
    flow.Func(func(v int) int { return v * 2 }),
    flow.Func(func(v int) int { return v - 3 }),
)
results := flow.Run(pipeline, 10) // [19]
```

### Chaining different types

Use `Pipe` to connect tasks with different input/output types:

```go
double := flow.Func(func(v int) int { return v * 2 })
toString := flow.Func(func(v int) string { return fmt.Sprintf("%d", v) })

pipeline := flow.Pipe(double, toString)
results := flow.Run(pipeline, 1, 2, 3) // ["2", "4", "6"]
```

For longer chains, nest `Pipe` calls:

```go
pipeline := flow.Pipe(flow.Pipe(parse, transform), encode)
```

### Concurrent workers

Fan out a function across multiple goroutines:

```go
heavy := flow.Concurrent(func(v int) int {
    time.Sleep(time.Second)
    return v * 2
}, 4)

results := flow.Run(heavy, 1, 2, 3, 4) // processed by 4 workers in parallel
```

Concurrent tasks can be composed with `Pipe` and `Chain` like any other task.
