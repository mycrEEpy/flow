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
				logger.Info(msg, args...)
			}

			out <- v
		}

		return nil
	}
}

// Filter returns a Task that only forwards items for which predicate returns true.
func Filter[T any](predicate func(T) bool) Task[T, T] {
	return func(in <-chan T, out chan<- T) error {
		defer close(out)

		for v := range in {
			if predicate(v) {
				out <- v
			}
		}

		return nil
	}
}
