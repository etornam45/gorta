package oauth

import (
	"context"
	"sync"
	"time"
)

const defaultStateTTL = 10 * time.Minute

type MemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]OAuthState
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		states: make(map[string]OAuthState),
	}
}

func (s *MemoryStateStore) Save(_ context.Context, state OAuthState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state.Token] = state
	return nil
}

func (s *MemoryStateStore) Find(_ context.Context, token string) (*OAuthState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[token]
	if !ok {
		return nil, ErrOAuthStateNotFound
	}
	if time.Now().After(state.ExpiresAt) {
		delete(s.states, token)
		return nil, ErrOAuthStateNotFound
	}
	return &state, nil
}

func (s *MemoryStateStore) Delete(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, token)
	return nil
}
