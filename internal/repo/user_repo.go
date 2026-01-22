package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
)

// UserRepository handles user data access
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, password_hash, role, supplier_id, status, created_at, updated_at 
              FROM users WHERE id = ?`

	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, password_hash, role, supplier_id, status, created_at, updated_at 
              FROM users WHERE username = ?`

	if err := r.db.GetContext(ctx, &user, query, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, pagination Pagination) (PaginatedResult[domain.User], error) {
	var users []domain.User
	var total int64

	// Count total
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM users"); err != nil {
		return PaginatedResult[domain.User]{}, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	query := `SELECT id, username, password_hash, role, supplier_id, status, created_at, updated_at 
              FROM users ORDER BY id LIMIT ? OFFSET ?`

	if err := r.db.SelectContext(ctx, &users, query, pagination.Limit(), pagination.Offset()); err != nil {
		return PaginatedResult[domain.User]{}, fmt.Errorf("failed to list users: %w", err)
	}

	return NewPaginatedResult(users, total, pagination), nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, password_hash, role, supplier_id, status) 
              VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.Role, user.SupplierID, user.Status)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	user.ID = uint64(id)

	return nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET username = ?, role = ?, supplier_id = ?, status = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, user.Username, user.Role, user.SupplierID, user.Status, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	query := `UPDATE users SET password_hash = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListBySupplier retrieves users by supplier ID
func (r *UserRepository) ListBySupplier(ctx context.Context, supplierID uint64) ([]domain.User, error) {
	var users []domain.User
	query := `SELECT id, username, password_hash, role, supplier_id, status, created_at, updated_at 
              FROM users WHERE supplier_id = ?`

	if err := r.db.SelectContext(ctx, &users, query, supplierID); err != nil {
		return nil, fmt.Errorf("failed to list users by supplier: %w", err)
	}

	return users, nil
}

// ExistsByUsername checks if a username exists
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = ?`

	if err := r.db.GetContext(ctx, &count, query, username); err != nil {
		return false, fmt.Errorf("failed to check username exists: %w", err)
	}

	return count > 0, nil
}
