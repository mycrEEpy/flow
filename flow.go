package flow

import "sync"

// Task processes values from in and sends results to out.
type Task[In, Out any] func(in <-chan In, out chan<- Out)

// TaskFunc is a function that converts a single value to a result.
type TaskFunc[In, Out any] func(In) Out

// Func creates a Task from a TaskFunc.
func Func[In, Out any](fn TaskFunc[In, Out]) Task[In, Out] {
	return func(in <-chan In, out chan<- Out) {
		for v := range in {
			out <- fn(v)
		}

		close(out)
	}
}

// Concurrent creates a Task that runs n goroutines applying fn concurrently.
func Concurrent[In, Out any](fn TaskFunc[In, Out], n int) Task[In, Out] {
	return func(in <-chan In, out chan<- Out) {
		var wg sync.WaitGroup
		wg.Add(n)

		for range n {
			go func() {
				defer wg.Done()
				for v := range in {
					out <- fn(v)
				}
			}()
		}

		wg.Wait()
		close(out)
	}
}

// Pipe connects two tasks: the output of a feeds into b.
func Pipe[A, B, C any](a Task[A, B], b Task[B, C]) Task[A, C] {
	return func(in <-chan A, out chan<- C) {
		ch := make(chan B)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			b(ch, out)
		}()

		a(in, ch)
		wg.Wait()
	}
}

// Chain composes multiple same-type tasks into a single task.
func Chain[T any](tasks ...Task[T, T]) Task[T, T] {
	t := tasks[0]

	for _, next := range tasks[1:] {
		t = Pipe(t, next)
	}

	return t
}

// Run starts the task with the given input values and returns the collected output.
func Run[In, Out any](t Task[In, Out], input ...In) []Out {
	in := make(chan In, len(input))

	for _, v := range input {
		in <- v
	}

	close(in)

	out := make(chan Out)

	var results []Out

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for v := range out {
			results = append(results, v)
		}
	}()

	t(in, out)
	wg.Wait()

	return results
}
