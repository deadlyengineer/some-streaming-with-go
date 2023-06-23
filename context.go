package some-streaming-with-go

import "context"

// contextDone returns true if ctx.Err() != nil.
func contextDone(ctx context.Context) bool {
	return ctx.Err() != nil
}
