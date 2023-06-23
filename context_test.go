package some-streaming-with-go

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func TestContextDone(t *testing.T) {
	is := is.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	is.True(!contextDone(ctx))

	cancel()
	is.True(contextDone(ctx))
}
