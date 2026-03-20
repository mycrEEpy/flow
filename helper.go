package flow

// Logger is the interface used by logging helpers to emit structured log messages.
type Logger interface {
	Info(msg string, args ...any)
}

// LogEveryN returns a pass-through Task that logs a message every n items processed.
func LogEveryN[T any](n int, logger Logger, msg string, args ...any) Task[T, T] {
	return func(in <-chan T, out chan<- T) error {
		defer close(out)

		var count int

		for v := range in {
			count++

			if count%n == 0 {
				logger.Info(msg, append([]any{"count", count}, args...)...)
			}

			out <- v
		}

		logger.Info(msg, append([]any{"count", count}, args...)...)

		return nil
	}
}

// Filter returns a Task that only forwards items for which predicate returns true.
// The predicate may also return an error, which will be returned by the Task to abort the pipeline.
func Filter[T any](predicate func(T) (bool, error)) Task[T, T] {
	return func(in <-chan T, out chan<- T) error {
		defer close(out)

		for v := range in {
			ok, err := predicate(v)
			if err != nil {
				return err
			}

			if ok {
				out <- v
			}
		}

		return nil
	}
}

// ForEach returns a Task that calls fn for each item and forwards the result.
// The function may also return an error, which will be returned by the Task to abort the pipeline.
func ForEach[In, Out any](fn func(in In) (Out, error)) Task[In, Out] {
	return func(in <-chan In, out chan<- Out) error {
		defer close(out)

		for v := range in {
			result, err := fn(v)
			if err != nil {
				return err
			}

			out <- result
		}

		return nil
	}
}

// Append returns a Task that appends all items to the given slice.
func Append[T any](s *[]T) Task[T, T] {
	return func(in <-chan T, out chan<- T) error {
		defer close(out)

		for v := range in {
			*s = append(*s, v)
			out <- v
		}

		return nil
	}
}

// Tee returns a Task that forwards all items to the given channel and the output channel.
func Tee[T any](tee chan<- T) Task[T, T] {
	return func(in <-chan T, out chan<- T) error {
		defer close(out)

		for v := range in {
			tee <- v
			out <- v
		}

		return nil
	}
}
