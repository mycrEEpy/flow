package flow_test

import (
	"errors"
	"testing"

	"github.com/mycreepy/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	double = func(in <-chan int, out chan<- int) error {
		defer close(out)
		for val := range in {
			out <- val * 2
		}
		return nil
	}

	triple = func(in <-chan int, out chan<- int) error {
		defer close(out)
		for val := range in {
			out <- val * 3
		}
		return nil
	}

	plusone = func(in <-chan int, out chan<- int) error {
		defer close(out)
		for val := range in {
			out <- val + 1
		}
		return nil
	}

	half = func(in <-chan int, out chan<- float64) error {
		defer close(out)
		for val := range in {
			out <- float64(val) / 2
		}
		return nil
	}
)

func TestSimpleTask(t *testing.T) {
	results, err := flow.FromValues(double, 1, 2, 3, 4, 5)

	require.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6, 8, 10}, results)
}

func TestConcurrentTask(t *testing.T) {
	results, err := flow.FromValues(flow.Concurrent(4, double), 1, 2, 3, 4, 5)

	require.NoError(t, err)
	assert.Contains(t, results, 2)
	assert.Contains(t, results, 4)
	assert.Contains(t, results, 6)
	assert.Contains(t, results, 8)
	assert.Contains(t, results, 10)
}

func TestPipe(t *testing.T) {
	task := flow.Pipe(plusone, half)

	results, err := flow.FromValues(task, 1, 2, 3, 4, 5)

	require.NoError(t, err)
	assert.Equal(t, []float64{1, 1.5, 2, 2.5, 3}, results)
}

func TestChain(t *testing.T) {
	task := flow.Chain(double, triple, double, triple)

	result, err := flow.FromValues(task, 1, 2, 3, 4, 5)

	require.NoError(t, err)
	assert.Equal(t, []int{36, 72, 108, 144, 180}, result)
}

func TestFromChannel(t *testing.T) {
	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := range 5 {
			ch <- i
		}
	}()

	result, err := flow.FromChannel(double, ch)

	require.NoError(t, err)
	assert.Equal(t, []int{0, 2, 4, 6, 8}, result)
}

func TestTaskError(t *testing.T) {
	errBoom := errors.New("boom")

	failing := func(in <-chan int, out chan<- int) error {
		defer close(out)
		for range in {
			return errBoom
		}
		return nil
	}

	_, err := flow.FromValues(failing, 1, 2, 3)

	assert.ErrorIs(t, err, errBoom)
}

func TestPipeErrorInFirstTask(t *testing.T) {
	errBoom := errors.New("boom")

	failing := func(in <-chan int, out chan<- int) error {
		defer close(out)
		for range in {
			return errBoom
		}
		return nil
	}

	task := flow.Pipe(failing, half)

	_, err := flow.FromValues(task, 1, 2, 3)

	assert.ErrorIs(t, err, errBoom)
}
