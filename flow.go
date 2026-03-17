package flow

import "sync"

// Task processes values from in and sends results to out.
// Must close out when done, including when returning an error.
// Returning a non-nil error signals that processing has failed;
// closing the out channel ensures downstream tasks and goroutines terminate cleanly.
type Task[In, Out any] func(in <-chan In, out chan<- Out) error

// Concurrent runs n goroutines of the given task sharing the same in/out channels.
// If any worker returns an error, the first non-nil error is returned.
func Concurrent[In, Out any](n int, t Task[In, Out]) Task[In, Out] {
	return func(in <-chan In, out chan<- Out) error {
		var (
			wg   sync.WaitGroup
			once sync.Once
			errs = make(chan error, n)
		)

		wg.Add(n)

		for range n {
			ch := make(chan Out)

			go func() {
				errs <- t(in, ch)
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

		var firstErr error

		for range n {
			if err := <-errs; err != nil {
				once.Do(func() { firstErr = err })
			}
		}

		return firstErr
	}
}

// Pipe connects two tasks: the output of a feeds into b.
// If either task returns an error, it is propagated to the caller.
func Pipe[A, B, C any](a Task[A, B], b Task[B, C]) Task[A, C] {
	return func(in <-chan A, out chan<- C) error {
		ch := make(chan B)

		var (
			wg   sync.WaitGroup
			bErr error
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			bErr = b(ch, out)
		}()

		aErr := a(in, ch)
		wg.Wait()

		if aErr != nil {
			return aErr
		}

		return bErr
	}
}

// Chain composes multiple same-type tasks into a single task.
// If any task in the chain returns an error, it is propagated to the caller.
func Chain[T any](tasks ...Task[T, T]) Task[T, T] {
	t := tasks[0]

	for _, next := range tasks[1:] {
		t = Pipe(t, next)
	}

	return t
}

// FromValues starts the task with the given input values and returns the collected output.
// Returns a non-nil error if the task fails.
func FromValues[In, Out any](t Task[In, Out], input ...In) ([]Out, error) {
	return FromSlice(t, input)
}

// FromSlice starts the task with the given input slice and returns the collected output.
// Returns a non-nil error if the task fails.
func FromSlice[In, Out any](t Task[In, Out], input []In) ([]Out, error) {
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
// Returns a non-nil error if the task fails.
func FromChannel[In, Out any](t Task[In, Out], in <-chan In) ([]Out, error) {
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

	err := t(in, out)
	wg.Wait()

	return results, err
}
