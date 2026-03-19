package flow_test

import (
	"testing"

	"github.com/mycreepy/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogger struct {
	calls []string
}

func (l *mockLogger) Info(msg string, args ...any) {
	l.calls = append(l.calls, msg)
}

func TestLogEveryN(t *testing.T) {
	logger := &mockLogger{}
	task := flow.LogEveryN[int](3, logger, "progress")

	results, err := flow.FromValues(task, 1, 2, 3, 4, 5, 6)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, results)
	assert.Equal(t, []string{"progress", "progress"}, logger.calls)
}

func TestFilter(t *testing.T) {
	task := flow.Filter(func(v int) bool { return v%2 == 0 })

	results, err := flow.FromValues(task, 1, 2, 3, 4, 5, 6)

	require.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6}, results)
}

func TestFilterNoneMatch(t *testing.T) {
	task := flow.Filter(func(v int) bool { return v > 10 })

	results, err := flow.FromValues(task, 1, 2, 3)

	require.NoError(t, err)
	assert.Empty(t, results)
}
