package bots

type Signal[T any] struct {
	Value T

	signal chan T
}

func NewSignal[T any]() *Signal[T] {
	return &Signal[T]{
		signal: make(chan T),
	}
}

func (s *Signal[T]) C() chan T {
	return s.signal
}

func (s *Signal[T]) Close() {
	if s.signal != nil {
		close(s.signal)
		s.signal = nil
	}
}
