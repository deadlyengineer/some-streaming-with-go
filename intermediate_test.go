package gostreams

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/matryer/is"
	"golang.org/x/exp/slices"
)

func TestMap(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = Map(ints, func(_ context.Context, _ context.CancelCauseFunc, elem int, _ uint64) int {
		return elem * 2
	})

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{2, 4, 6, 8, 10})
}

func TestMap_Cancel(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = Map(ints, func(_ context.Context, cancel context.CancelCauseFunc, elem int, _ uint64) int {
		is.True(elem <= 3)

		if elem == 3 {
			cancel(nil)
			return 0
		}

		return elem * 2
	})

	result, err := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{2, 4})
	is.True(errors.Is(err, context.Canceled))
}

func TestMapConcurrent(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = MapConcurrent(ints, func(_ context.Context, _ context.CancelCauseFunc, elem int, _ uint64) int {
		return elem * 2
	})

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	slices.Sort(result)

	is.Equal(result, []int{2, 4, 6, 8, 10})
}

func TestFilter(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = Filter(ints, even)

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{2, 4})
}

func TestFilter_Cancel(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	evenCancel := func(_ context.Context, cancel context.CancelCauseFunc, elem int, _ uint64) bool {
		is.True(elem <= 3)

		if elem == 3 {
			cancel(nil)
			return false
		}

		return elem%2 == 0
	}

	ints = Filter(ints, evenCancel)

	result, err := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{2})
	is.True(errors.Is(err, context.Canceled))
}

func TestFilterConcurrent(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = FilterConcurrent(ints, even)

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	slices.Sort(result)

	is.Equal(result, []int{2, 4})
}

func TestPeek(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	sum := 0

	ints = Peek(ints, func(_ context.Context, _ context.CancelCauseFunc, elem int, _ uint64) {
		sum += elem
	})

	_, _ = Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(sum, 15)
}

func TestPeek_Cancel(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	sum := 0

	ints = Peek(ints, func(_ context.Context, cancel context.CancelCauseFunc, elem int, _ uint64) {
		is.True(elem <= 3)

		if elem == 3 {
			cancel(nil)
			return
		}

		sum += elem
	})

	_, err := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(sum, 3)
	is.True(errors.Is(err, context.Canceled))
}

func TestSort(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{3, 1, 2, 4, 5})

	ints = Sort(ints, func(_ context.Context, _ context.CancelCauseFunc, a int, b int) bool {
		return a < b
	})

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{1, 2, 3, 4, 5})
}

func TestSort_Cancel(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{3, 1, 2, 4, 5})

	ints = Sort(ints, func(_ context.Context, cancel context.CancelCauseFunc, a int, b int) bool { //nolint:varnamelen // a and b are okay for sorting
		if a == 4 || b == 4 {
			cancel(nil)
			return false
		}

		return a < b
	})

	_, err := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.True(errors.Is(err, context.Canceled))
}

func TestLimit(t *testing.T) { //nolint:gocognit // it's a bit more involved
	tests := []struct {
		givenLimit           uint64
		want                 []int
		wantProducerCanceled bool
	}{
		{
			givenLimit:           3,
			want:                 []int{1, 2, 2, 3, 3, 3},
			wantProducerCanceled: true,
		},
		{
			givenLimit:           0,
			want:                 nil,
			wantProducerCanceled: true,
		},
		{
			givenLimit:           100,
			want:                 []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 5, 1, 2, 2, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 5, 1, 2, 2, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 5},
			wantProducerCanceled: false,
		},
	}

	for idx, test := range tests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			is := is.New(t)

			ctx := context.Background()

			producerCanceled := make(chan bool)

			ints := func(ctx context.Context, _ context.CancelCauseFunc) <-chan int {
				outCh := make(chan int)

				go func() {
					canceled := false

					defer func() {
						producerCanceled <- canceled
					}()

					defer close(outCh)

					for _, i := range []int{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5} {
						select {
						case outCh <- i:

						case <-ctx.Done():
							canceled = true
							return
						}
					}
				}()

				return outCh
			}

			ints = Limit(ints, test.givenLimit)

			ints = FlatMap(ints, func(_ context.Context, _ context.CancelCauseFunc, elem int, _ uint64) ProducerFunc[int] {
				elems := make([]int, elem)
				for i := 0; i < elem; i++ {
					elems[i] = elem
				}

				return Produce(elems)
			})

			result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

			is.Equal(result, test.want)
			is.Equal(<-producerCanceled, test.wantProducerCanceled)
		})
	}
}

func TestSkip(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = Skip(ints, 3)

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{4, 5})
}

func TestFlatMap(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = FlatMap(ints, func(_ context.Context, _ context.CancelCauseFunc, elem int, _ uint64) ProducerFunc[int] {
		elems := make([]int, elem)
		for i := 0; i < elem; i++ {
			elems[i] = i + 1
		}

		return Produce(elems)
	})

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	is.Equal(result, []int{1, 1, 2, 1, 2, 3, 1, 2, 3, 4, 1, 2, 3, 4, 5})
}

func TestFlatMapConcurrent(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	ints := Produce([]int{1, 2, 3, 4, 5})

	ints = FlatMapConcurrent(ints, func(_ context.Context, _ context.CancelCauseFunc, elem int, _ uint64) ProducerFunc[int] {
		elems := make([]int, elem)
		for i := 0; i < elem; i++ {
			elems[i] = i + 1
		}

		return Produce(elems)
	})

	result, _ := Reduce(ctx, ints, nil, CollectSlice[int]())

	slices.Sort(result)

	is.Equal(result, []int{1, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 4, 4, 5})
}