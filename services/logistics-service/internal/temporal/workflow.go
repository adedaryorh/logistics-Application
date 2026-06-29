package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderSagaInput struct {
	OrderID    string
	CustomerID string
	Type       string
}

type Coordinator struct{}

func NewCoordinator() *Coordinator {
	return &Coordinator{}
}

func (c *Coordinator) StartOrderSaga(ctx context.Context, input OrderSagaInput) (string, error) {
	workflowID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate workflow id: %w", err)
	}
	return "order-saga-" + workflowID.String(), nil
}

func OrderSagaWorkflow(ctx context.Context, input OrderSagaInput) error {
	_ = []string{
		"ValidateOrder",
		"CalculateFare",
		"InitializePayment",
		"WaitForPayment",
		"FindAndDispatch",
		"BeginTracking",
		"WaitForCompletion",
		"Settle",
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return nil
	}
}
