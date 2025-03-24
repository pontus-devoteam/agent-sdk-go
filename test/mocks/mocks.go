package mocks

import (
	"context"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/stretchr/testify/mock"
)

// MockModel is a mock implementation of the model.Model interface
type MockModel struct {
	mock.Mock
}

func (m *MockModel) GetResponse(ctx context.Context, request *model.Request) (*model.Response, error) {
	args := m.Called(ctx, request)
	if resp := args.Get(0); resp != nil {
		return resp.(*model.Response), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockModel) StreamResponse(ctx context.Context, request *model.Request) (<-chan model.StreamEvent, error) {
	args := m.Called(ctx, request)
	if ch := args.Get(0); ch != nil {
		return ch.(<-chan model.StreamEvent), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockModelProvider is a mock implementation of the model.Provider interface
type MockModelProvider struct {
	mock.Mock
}

func (p *MockModelProvider) GetModel(name string) (model.Model, error) {
	args := p.Called(name)
	if m := args.Get(0); m != nil {
		return m.(model.Model), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockStateStore is a mock implementation of the WorkflowStateStore interface
type MockStateStore struct {
	mock.Mock
}

func (s *MockStateStore) SaveState(workflowID string, state interface{}) error {
	args := s.Called(workflowID, state)
	return args.Error(0)
}

func (s *MockStateStore) LoadState(workflowID string) (interface{}, error) {
	args := s.Called(workflowID)
	return args.Get(0), args.Error(1)
}

func (s *MockStateStore) ListCheckpoints(workflowID string) ([]string, error) {
	args := s.Called(workflowID)
	return args.Get(0).([]string), args.Error(1)
}

func (s *MockStateStore) DeleteCheckpoint(workflowID string, checkpointID string) error {
	args := s.Called(workflowID, checkpointID)
	return args.Error(0)
}

// InMemoryStateStore implements a simple in-memory state store for testing
type InMemoryStateStore struct {
	states map[string]interface{}
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{
		states: make(map[string]interface{}),
	}
}

func (s *InMemoryStateStore) SaveState(workflowID string, state interface{}) error {
	s.states[workflowID] = state
	return nil
}

func (s *InMemoryStateStore) LoadState(workflowID string) (interface{}, error) {
	if state, exists := s.states[workflowID]; exists {
		return state, nil
	}
	return nil, nil
}

func (s *InMemoryStateStore) ListCheckpoints(workflowID string) ([]string, error) {
	checkpoints := make([]string, 0)
	if _, exists := s.states[workflowID]; exists {
		checkpoints = append(checkpoints, workflowID)
	}
	return checkpoints, nil
}

func (s *InMemoryStateStore) DeleteCheckpoint(workflowID string, checkpointID string) error {
	delete(s.states, workflowID)
	return nil
}
