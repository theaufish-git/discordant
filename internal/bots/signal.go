package bots

import (
	"sync"
)

type Signal[T any] struct {
	sync.RWMutex
	value T
	subs  []chan<- T
}

func NewSignal[T any]() *Signal[T] {
	return &Signal[T]{}
}

func (s *Signal[T]) Close() {
	for _, ch := range s.subs {
		close(ch)
	}
}

func (s *Signal[T]) Get() T {
	s.RLock()
	defer s.RUnlock()

	return s.value
}

func (s *Signal[T]) Set(v T) {
	s.Lock()
	s.value = v
	s.Unlock()

	for _, ch := range s.subs {
		ch <- v
	}
}

func (s *Signal[T]) Subscribe(ch chan<- T) {
	s.Lock()
	val := s.value
	s.subs = append(s.subs, ch)
	s.Unlock()
	go func() { ch <- val }()
}
