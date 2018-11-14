package license

import (
	"context"
)

// StatusType is the type of status message, such as normal, error, warning.
type StatusType uint

const (
	StatusUnknown StatusType = iota
	StatusNormal
	StatusWarning
	StatusError
)

// StatusListener is called to update the status of a finder.
type StatusListener interface {
	// UpdateStatus is called whenever there is an updated status message.
	// This function must not block, since this will block the actual
	// behavior of the license finder as well. If blocking behavior is
	// necessary, end users should use a channel internally to avoid it on
	// the function call.
	//
	// The message should be relatively short if possible (within ~50 chars)
	// so that it fits nicely on a terminal. It should be a basic status
	// update.
	UpdateStatus(t StatusType, msg string)
}

// StatusWithContext inserts a StatusListener into a context.
func StatusWithContext(ctx context.Context, l StatusListener) context.Context {
	return context.WithValue(ctx, statusCtxKey, l)
}

// UpdateStatus updates the status of the listener (if any) in the given
// context.
func UpdateStatus(ctx context.Context, t StatusType, msg string) {
	sl, ok := ctx.Value(statusCtxKey).(StatusListener)
	if !ok || sl == nil {
		return
	}

	sl.UpdateStatus(t, msg)
}

type statusCtxKeyType struct{}

var statusCtxKey = statusCtxKeyType{}
