package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber2/app/model"
)

var ErrNotFound = errors.New("data tidak ditemukan")
var ErrDuplicate = errors.New("data sudah ada")

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	Create(ctx context.Context, user model.User) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at 
		 FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at 
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	return u, nil
}

func (r *userPostgresRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role, is_active) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		user.Username, user.Email, user.Password, user.Role, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return model.User{}, err
	}
	return user, nil
}