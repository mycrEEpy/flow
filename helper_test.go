package flow_test

import (
	"errors"
	"fmt"
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
	task := flow.LogEveryN[int](2, logger, "progress", "extra", "value")

	results, err := flow.FromValues(task, 1, 2, 3, 4, 5)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, results)
	require.Len(t, logger.entries, 3)
	assert.Equal(t, "progress", logger.entries[0].msg)
	assert.Equal(t, []any{"count", 2, "extra", "value"}, logger.entries[0].args)
	assert.Equal(t, []any{"count", 4, "extra", "value"}, logger.entries[1].args)
	assert.Equal(t, []any{"count", 5, "closed", true, "extra", "value"}, logger.entries[2].args)
}

func TestFilter(t *testing.T) {
	task := flow.Filter(func(v int) (bool, error) { return v%2 == 0, nil })

	results, err := flow.FromValues(task, 1, 2, 3, 4, 5, 6)

	require.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6}, results)
}

func TestFilterNoneMatch(t *testing.T) {
	task := flow.Filter(func(v int) (bool, error) { return v > 10, nil })

	results, err := flow.FromValues(task, 1, 2, 3)

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestFilterError(t *testing.T) {
	errBoom := errors.New("boom")

	task := flow.Filter(func(v int) (bool, error) {
		if v < 0 {
			return false, errBoom
		}
		return true, nil
	})

	_, err := flow.FromValues(task, 1, -1, 3)

	assert.ErrorIs(t, err, errBoom)
}

func TestForEach(t *testing.T) {
	task := flow.ForEach(func(v int) (int, error) { return v * 2, nil })

	results, err := flow.FromValues(task, 1, 2, 3)

	require.NoError(t, err)
	assert.Equal(t, []int{2, 4, 6}, results)
}

func TestForEachTypeChange(t *testing.T) {
	task := flow.ForEach(func(v int) (string, error) { return fmt.Sprintf("%d", v), nil })

	results, err := flow.FromValues(task, 1, 2, 3)

	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, results)
}

func TestForEachError(t *testing.T) {
	errBoom := errors.New("boom")

	task := flow.ForEach(func(v int) (int, error) {
		if v < 0 {
			return 0, errBoom
		}
		return v, nil
	})

	_, err := flow.FromValues(task, 1, -1, 3)

	assert.ErrorIs(t, err, errBoom)
}

func TestAppend(t *testing.T) {
	var collected []int
	task := flow.Append(&collected)

	results, err := flow.FromValues(task, 1, 2, 3)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, results)
	assert.Equal(t, []int{1, 2, 3}, collected)
}

func TestAppendEmpty(t *testing.T) {
	var collected []int
	task := flow.Append(&collected)

	results, err := flow.FromValues[int, int](task)

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Empty(t, collected)
}

func TestTeeEmpty(t *testing.T) {
	tee := make(chan int, 1)
	task := flow.Tee(tee)

	results, err := flow.FromValues[int, int](task)
	close(tee)

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Empty(t, tee)
}

func TestTee(t *testing.T) {
	tee := make(chan int, 3)
	task := flow.Tee(tee)

	results, err := flow.FromValues(task, 1, 2, 3)
	close(tee)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, results)

	var teeValues []int
	for v := range tee {
		teeValues = append(teeValues, v)
	}
	assert.Equal(t, []int{1, 2, 3}, teeValues)
}
