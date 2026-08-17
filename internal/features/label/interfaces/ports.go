package interfaces

import (
	"context"

	"github.com/sraj/addressbook/internal/features/label/domain"
)

// labelService is the consumer-facing port the handler depends on.
type labelService interface {
	GenerateSheet(ctx context.Context, userID, collectionID uint, format string) (string, error)
	CreateOrder(ctx context.Context, userID, collectionID uint, userEmail, format string) (*domain.LabelOrder, string, error)
	ListOrders(ctx context.Context, userID uint) ([]domain.LabelOrder, error)
	ConfirmOrder(ctx context.Context, userID uint, sessionID string) (*domain.LabelOrder, error)
	MarkOrderPaidBySessionID(ctx context.Context, sessionID string) error
	SupportedFormats() []domain.Format
}
