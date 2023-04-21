package gostreams

import (
	"context"
	"fmt"
	"strconv"
)

func Example() {
	// construct a producer from a slice
	ints := Produce([]int{1, 2, 3, 4, 5})

	// map elements by doubling them
	// since we only need the elements themselves, we can use FuncMapper
	ints = Map(ints, FuncMapper(func(elem int) int {
		return elem * 2
	}))

	// map elements by converting them to strings
	intStrs := Map(ints, FuncMapper(strconv.Itoa))

	// perform a reduction to collect the strings into a slice
	// since append can work with nil, we can simply pass nil as the initial accumulator
	strs, _ := Reduce(context.Background(), intStrs, nil, CollectSlice[string]())

	fmt.Printf("%+v\n", strs)
	// Output: [2 4 6 8 10]
}