package flow

import "sync"

// Task processes values from in and sends results to out.
// Must close out when done.
type Task[In, Out any] func(in <-chan In, out chan<- Out)

// Concurrent runs n goroutines of the given task sharing the same in/out channels.
func Concurrent[In, Out any](t Task[In, Out], n int) Task[In, Out] {
	return func(in <-chan In, out chan<- Out) {
		var wg sync.WaitGroup
		wg.Add(n)

		for range n {
			ch := make(chan Out)

			go func() {
				t(in, ch)
			}()

			go func() {
				defer wg.Done()
				for v := range ch {
					out <- v
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

// FromValues starts the task with the given input values and returns the collected output.
func FromValues[In, Out any](t Task[In, Out], input ...In) []Out {
	ch := make(chan In)

	go func() {
		defer close(ch)
		for _, v := range input {
			ch <- v
		}
	}()

	return FromChannel(t, ch)
}

// FromChannel starts the task with values from a channel and returns the collected output.
func FromChannel[In, Out any](t Task[In, Out], in <-chan In) []Out {
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
