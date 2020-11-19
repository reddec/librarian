package librarian

import "context"

// Storage for indexed data.
type Storage interface {
	// Get data by ID.
	Get(ctx context.Context, id string) ([]byte, error)
	// Iterate over all items.
	Iterate(ctx context.Context, iterator func(id string, data []byte) error) error
	// Delete item by id.
	Delete(ctx context.Context, id string) error
	// Update or replace item by id.
	Update(ctx context.Context, id string, data []byte) error
	// Create record and issue id.
	Create(ctx context.Context, data []byte) (string, error)
}
