[![GoDoc](https://pkg.go.dev/badge/github.com/deadlyengineer/some-streaming-with-go)](https://pkg.go.dev/github.com/deadlyengineer/some-streaming-with-go)


some-streaming-with-go
=========

A Go package that provides a set of operations on streams of elements.

```go
import "github.com/deadlyengineer/some-streaming-with-go"
```


Code example
------------

```go
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
strs, _ := ReduceSlice(context.Background(), intStrs)

fmt.Printf("%+v\n", strs)
// Output: [2 4 6 8 10]
```


License
-------

This package is licensed under the MIT license.
