package flow_test

import (
	"testing"

	"github.com/mycreepy/flow"
	"github.com/stretchr/testify/assert"
)

func TestFunc(t *testing.T) {
	double := flow.Func[int, int](func(in int) int {
		return in * 2
	})

	results := flow.Run(double, 1, 2, 3, 4, 5)

	assert.Equal(t, []int{2, 4, 6, 8, 10}, results)
}

func TestConcurrent(t *testing.T) {
	double := flow.Concurrent[int, int](func(in int) int {
		return in * 2
	}, 4)

	results := flow.Run(double, 1, 2, 3, 4, 5)

	assert.Contains(t, results, 2)
	assert.Contains(t, results, 4)
	assert.Contains(t, results, 6)
	assert.Contains(t, results, 8)
	assert.Contains(t, results, 10)
}

func TestPipe(t *testing.T) {
	plusone := flow.Func[int, int](func(in int) int {
		return in + 1
	})

	half := flow.Func[int, float64](func(in int) float64 {
		return float64(in) / 2
	})

	task := flow.Pipe(plusone, half)

	results := flow.Run(task, 1, 2, 3, 4, 5)

	assert.Equal(t, []float64{1, 1.5, 2, 2.5, 3}, results)
}

func TestChain(t *testing.T) {
	double := flow.Func[int, int](func(in int) int {
		return in * 2
	})

	triple := flow.Func[int, int](func(in int) int {
		return in * 3
	})

	task := flow.Chain(double, triple, double, triple)

	result := flow.Run(task, 1, 2, 3, 4, 5)

	assert.Equal(t, []int{36, 72, 108, 144, 180}, result)
}
