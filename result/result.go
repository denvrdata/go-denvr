package result

// https://en.wikipedia.org/wiki/Result_type
type Result[T any] struct {
	value T
	err   error
}

func Wrap[T any](value T, err error) Result[T] {
	return Result[T]{value, err}
}

func (r Result[T]) Unwrap() T {
	if r.Ok() {
		return r.value
	}
	panic(r.err)
}

func (r Result[T]) Ok() bool {
	return r.err == nil
}
