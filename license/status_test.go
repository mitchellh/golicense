package license

import (
	"context"
	"testing"
)

func TestUpdateStatus_noExist(t *testing.T) {
	// Just basically testing that we don't panic here
	UpdateStatus(context.Background(), StatusNormal, "hello")
}

func TestUpdateStatus_badType(t *testing.T) {
	// Just basically testing that we don't panic here
	UpdateStatus(context.WithValue(context.Background(), statusCtxKey, 42),
		StatusNormal, "hello")
}

func TestUpdateStatus_good(t *testing.T) {
	var mock MockStatusListener
	ctx := StatusWithContext(context.Background(), &mock)

	mock.On("UpdateStatus", StatusNormal, "hello").Once()
	UpdateStatus(ctx, StatusNormal, "hello")
	mock.AssertExpectations(t)

	mock.On("UpdateStatus", StatusWarning, "warning").Once()
	UpdateStatus(ctx, StatusWarning, "warning")
	mock.AssertExpectations(t)
}
