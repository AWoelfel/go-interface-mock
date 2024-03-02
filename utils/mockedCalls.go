package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type MockedCalls interface {
	Next(t *testing.T, expected string) (int, []interface{})
	AppendCall(name string, expectedObjects ...interface{})
	AssertNoCallsLeft(t *testing.T) bool
}

func NewMockedCalls() *mockedCalls {
	return &mockedCalls{}
}

type mockedCalls struct {
	sync.Mutex
	idx             int
	expectedObjects [][]interface{}
}

func (m *mockedCalls) AssertNoCallsLeft(t *testing.T) bool {
	t.Helper()

	if len(m.expectedObjects) != m.idx {
		missingCalls := []string{}

		for _, objs := range m.expectedObjects[m.idx:] {
			missingCalls = append(missingCalls, objs[0].(string))
		}

		return assert.EqualValuesf(t, len(m.expectedObjects), m.idx, "calls which where not done : %v", missingCalls)
	}

	return true
}

func (m *mockedCalls) Next(t *testing.T, expected string) (int, []interface{}) {
	m.Lock()
	defer m.Unlock()
	t.Helper()

	require.Truef(t, m.idx < len(m.expectedObjects), "no more calls expected, expected number of calls :%d, unexpected call of %q occurred", len(m.expectedObjects), expected)

	objects := m.expectedObjects[m.idx]
	require.EqualValuesf(t, objects[0], expected, "Method mismatch on call #%d", m.idx)

	defer func() { m.idx++ }()

	return m.idx, objects[1:]
}

func (m *mockedCalls) AppendCall(name string, expectedObjects ...interface{}) {
	m.Lock()
	defer m.Unlock()

	objectsToAdd := []interface{}{name}
	objectsToAdd = append(objectsToAdd, expectedObjects...)

	m.expectedObjects = append(m.expectedObjects, objectsToAdd)
}
