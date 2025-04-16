package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateUser(ctx context.Context, req SignupRequest) (*User, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if count > 0 {
		return nil, ErrEmailTaken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	user := User{
		ID:             uuid.New(),
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
		DisplayName:    req.DisplayName,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, hashed_password, display_name, created_at, updated_at, disabled) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Email, user.HashedPassword, user.DisplayName, user.CreatedAt, user.UpdatedAt, false)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*User, *Session, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, hashed_password, display_name, created_at, updated_at, last_login, disabled 
         FROM users WHERE email = $1`,
		req.Email).Scan(
		&user.ID, &user.Email, &user.HashedPassword, &user.DisplayName,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.Disabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, fmt.Errorf("error fetching user: %w", err)
	}

	if user.Disabled {
		return nil, nil, errors.New("account is disabled")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	session := Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (id, user_id, token, expires_at, created_at) 
         VALUES ($1, $2, $3, $4, $5)`,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating session: %w", err)
	}

	now := time.Now()
	user.LastLogin = &now
	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET last_login = $1 WHERE id = $2",
		now, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("error updating last login: %w", err)
	}

	return &user, &session, nil
}

func (s *Service) GetUserByToken(ctx context.Context, token string) (*User, error) {
	var sessionUserID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		"SELECT user_id FROM sessions WHERE token = $1 AND expires_at > $2",
		token, time.Now()).Scan(&sessionUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("error fetching session: %w", err)
	}

	var user User
	err = s.db.QueryRowContext(ctx,
		`SELECT id, email, display_name, created_at, updated_at, last_login, disabled 
         FROM users WHERE id = $1 AND NOT disabled`,
		sessionUserID).Scan(
		&user.ID, &user.Email, &user.DisplayName,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.Disabled)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return &user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rows == 0 {
		return ErrInvalidToken
	}

	return nil
}
