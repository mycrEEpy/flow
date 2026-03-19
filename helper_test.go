package flow_test

import (
	"testing"

	"github.com/mycreepy/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logEntry struct {
	msg  string
	args []any
}

type mockLogger struct {
	entries []logEntry
}

func (l *mockLogger) Info(msg string, args ...any) {
	l.entries = append(l.entries, logEntry{msg, args})
}

func TestLogEveryN(t *testing.T) {
	logger := &mockLogger{}
	task := flow.LogEveryN[int](3, logger, "progress", "extra", "value")

	results, err := flow.FromValues(task, 1, 2, 3, 4, 5, 6)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, results)
	require.Len(t, logger.entries, 2)
	assert.Equal(t, "progress", logger.entries[0].msg)
	assert.Equal(t, []any{"count", 3, "extra", "value"}, logger.entries[0].args)
	assert.Equal(t, []any{"count", 6, "extra", "value"}, logger.entries[1].args)
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
