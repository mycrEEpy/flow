package flow_test

import (
	"testing"

	"github.com/mycreepy/flow"
	"github.com/stretchr/testify/assert"
)

var (
	double = func(in <-chan int, out chan<- int) {
		defer close(out)
		for val := range in {
			out <- val * 2
		}
	}

	triple = func(in <-chan int, out chan<- int) {
		defer close(out)
		for val := range in {
			out <- val * 3
		}
	}

	plusone = func(in <-chan int, out chan<- int) {
		defer close(out)
		for val := range in {
			out <- val + 1
		}
	}

	half = func(in <-chan int, out chan<- float64) {
		defer close(out)
		for val := range in {
			out <- float64(val) / 2
		}
	}
)

func TestDo(t *testing.T) {
	results := flow.FromValues(double, 1, 2, 3, 4, 5)

	assert.Equal(t, []int{2, 4, 6, 8, 10}, results)
}

func TestConcurrent(t *testing.T) {
	results := flow.FromValues(flow.Concurrent(double, 4), 1, 2, 3, 4, 5)

	assert.Contains(t, results, 2)
	assert.Contains(t, results, 4)
	assert.Contains(t, results, 6)
	assert.Contains(t, results, 8)
	assert.Contains(t, results, 10)
}

func TestPipe(t *testing.T) {
	task := flow.Pipe(plusone, half)

	results := flow.FromValues(task, 1, 2, 3, 4, 5)

	assert.Equal(t, []float64{1, 1.5, 2, 2.5, 3}, results)
}

func TestChain(t *testing.T) {
	task := flow.Chain(double, triple, double, triple)

	result := flow.FromValues(task, 1, 2, 3, 4, 5)

	assert.Equal(t, []int{36, 72, 108, 144, 180}, result)
}

func TestRunWithProducer(t *testing.T) {
	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := range 5 {
			ch <- i
		}
	}()

	result := flow.FromChannel(double, ch)

	assert.Equal(t, []int{0, 2, 4, 6, 8}, result)
}
