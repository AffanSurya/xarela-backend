package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultBaseCurrency    = "IDR"
	defaultAccessTokenTTL  = time.Hour
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
	defaultVerificationTTL = 24 * time.Hour
)

var (
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrEmailVerificationInvalid = errors.New("email verification invalid or expired")
	ErrEmailVerificationExpired = errors.New("email verification expired")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrRefreshTokenInvalid      = errors.New("refresh token invalid or expired")
	ErrRefreshTokenReuse        = errors.New("refresh token reuse detected")
	ErrUnauthorized             = errors.New("unauthorized")
)

type AuthService struct {
	dsn                  string
	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
	verificationTokenTTL time.Duration
	now                  func() time.Time
}

type AuthUser struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	FullName        string `json:"full_name"`
	IsEmailVerified bool   `json:"is_email_verified"`
}

type AuthTokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type AuthEnvelope struct {
	User  *AuthUser     `json:"user,omitempty"`
	Token AuthTokenPair `json:"-"`
}

type RegisterRequest struct {
	Email    string
	Password string
	FullName string
}

type LoginRequest struct {
	Email    string
	Password string
}

type RefreshRequest struct {
	RefreshToken string
}

type VerifyEmailRequest struct {
	VerificationToken string
}

type RegisterResult struct {
	User  AuthUser
	Token AuthTokenPair
}

type LoginResult struct {
	Token AuthTokenPair
}

type RefreshResult struct {
	Token AuthTokenPair
}

type LogoutResult struct {
	Revoked bool `json:"revoked"`
}

type VerifyEmailResult struct {
	Verified bool `json:"verified"`
}

type ValidationError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewAuthService(dsn string) AuthService {
	return AuthService{
		dsn:                  strings.TrimSpace(dsn),
		accessTokenTTL:       defaultAccessTokenTTL,
		refreshTokenTTL:      defaultRefreshTokenTTL,
		verificationTokenTTL: defaultVerificationTTL,
		now:                  time.Now,
	}
}

func (s AuthService) Register(ctx context.Context, request RegisterRequest) (RegisterResult, error) {
	if validationErr := validateRegisterRequest(request); validationErr != nil {
		return RegisterResult{}, validationErr
	}

	database, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return RegisterResult{}, fmt.Errorf("ping database: %w", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	email := normalizeEmail(request.Email)
	if exists, err := emailExists(ctx, tx, email); err != nil {
		return RegisterResult{}, err
	} else if exists {
		return RegisterResult{}, &ValidationError{
			Code:    "email_already_exists",
			Message: "Email already registered",
			Fields: map[string]string{
				"email": "duplicate",
			},
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hash password: %w", err)
	}

	verificationToken, err := generateToken(32)
	if err != nil {
		return RegisterResult{}, err
	}
	verificationExpiry := s.now().Add(s.verificationTokenTTL)

	userID := uuid.New()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO users (id, email, password_hash, full_name, is_email_verified, email_verification_token_hash, email_verification_expires_at, base_currency)
VALUES ($1, $2, $3, $4, false, $5, $6, $7)
`, userID, email, string(hash), strings.TrimSpace(request.FullName), hashVerificationToken(verificationToken), verificationExpiry, defaultBaseCurrency); err != nil {
		return RegisterResult{}, fmt.Errorf("insert user: %w", err)
	}

	token, err := createSession(ctx, tx, s.now(), userID, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return RegisterResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return RegisterResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	slog.InfoContext(ctx, "email verification token generated", "email", email, "token", verificationToken)

	return RegisterResult{
		User: AuthUser{
			ID:              userID.String(),
			Email:           email,
			FullName:        strings.TrimSpace(request.FullName),
			IsEmailVerified: false,
		},
		Token: token,
	}, nil
}

func (s AuthService) VerifyEmail(ctx context.Context, request VerifyEmailRequest) (VerifyEmailResult, error) {
	if validationErr := validateVerifyEmailRequest(request); validationErr != nil {
		return VerifyEmailResult{}, validationErr
	}

	database, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return VerifyEmailResult{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return VerifyEmailResult{}, fmt.Errorf("ping database: %w", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return VerifyEmailResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	verificationTokenHash := hashVerificationToken(request.VerificationToken)
	var (
		userID    uuid.UUID
		verified  bool
		expiresAt time.Time
	)
	err = tx.QueryRowContext(ctx, `
SELECT id, is_email_verified, email_verification_expires_at
FROM users
WHERE email_verification_token_hash = $1
`, verificationTokenHash).Scan(&userID, &verified, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return VerifyEmailResult{}, ErrEmailVerificationInvalid
		}
		return VerifyEmailResult{}, err
	}

	now := s.now()
	if now.After(expiresAt) {
		if _, err := tx.ExecContext(ctx, `
UPDATE users
SET email_verification_token_hash = NULL,
    email_verification_expires_at = NULL
WHERE id = $1
`, userID); err != nil {
			return VerifyEmailResult{}, fmt.Errorf("clear expired verification token: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return VerifyEmailResult{}, fmt.Errorf("commit transaction: %w", err)
		}
		return VerifyEmailResult{}, ErrEmailVerificationExpired
	}

	if !verified {
		if _, err := tx.ExecContext(ctx, `
UPDATE users
SET is_email_verified = true,
    email_verification_token_hash = NULL,
    email_verification_expires_at = NULL
WHERE id = $1
`, userID); err != nil {
			return VerifyEmailResult{}, fmt.Errorf("verify user: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return VerifyEmailResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return VerifyEmailResult{Verified: true}, nil
}

func (s AuthService) Login(ctx context.Context, request LoginRequest) (LoginResult, error) {
	if validationErr := validateLoginRequest(request); validationErr != nil {
		return LoginResult{}, validationErr
	}

	database, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return LoginResult{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return LoginResult{}, fmt.Errorf("ping database: %w", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return LoginResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	userID, passwordHash, err := loadUserCredentials(ctx, tx, normalizeEmail(request.Email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := createSession(ctx, tx, s.now(), userID, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return LoginResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return LoginResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return LoginResult{Token: token}, nil
}

func (s AuthService) Refresh(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
	if validationErr := validateRefreshRequest(request); validationErr != nil {
		return RefreshResult{}, validationErr
	}

	database, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return RefreshResult{}, fmt.Errorf("ping database: %w", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	hash := hashRefreshToken(request.RefreshToken)
	session, err := loadSessionByHash(ctx, tx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if revokedUserID, revoked := lookupRevokedSession(ctx, tx, hash); revoked {
				if err := revokeAllUserSessions(ctx, tx, revokedUserID); err != nil {
					return RefreshResult{}, err
				}
				if err := tx.Commit(); err != nil {
					return RefreshResult{}, fmt.Errorf("commit transaction: %w", err)
				}
				return RefreshResult{}, ErrRefreshTokenReuse
			}
			return RefreshResult{}, ErrRefreshTokenInvalid
		}
		return RefreshResult{}, err
	}

	if !session.RevokedAt.IsZero() || s.now().After(session.ExpiresAt) {
		if !session.RevokedAt.IsZero() {
			if err := revokeAllUserSessions(ctx, tx, session.UserID); err != nil {
				return RefreshResult{}, err
			}
			if err := tx.Commit(); err != nil {
				return RefreshResult{}, fmt.Errorf("commit transaction: %w", err)
			}
			return RefreshResult{}, ErrRefreshTokenReuse
		}
		if err := revokeSessionByID(ctx, tx, session.ID); err != nil {
			return RefreshResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return RefreshResult{}, fmt.Errorf("commit transaction: %w", err)
		}
		return RefreshResult{}, ErrRefreshTokenInvalid
	}

	if err := revokeSessionByID(ctx, tx, session.ID); err != nil {
		return RefreshResult{}, err
	}

	token, err := createSession(ctx, tx, s.now(), session.UserID, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		return RefreshResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return RefreshResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return RefreshResult{Token: token}, nil
}

func (s AuthService) Logout(ctx context.Context, request RefreshRequest) (LogoutResult, error) {
	if validationErr := validateRefreshRequest(request); validationErr != nil {
		return LogoutResult{}, validationErr
	}

	database, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return LogoutResult{}, fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return LogoutResult{}, fmt.Errorf("ping database: %w", err)
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return LogoutResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	hash := hashRefreshToken(request.RefreshToken)
	session, err := loadSessionByHashAllowRevoked(ctx, tx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LogoutResult{}, ErrUnauthorized
		}
		return LogoutResult{}, err
	}

	if !session.RevokedAt.IsZero() {
		if err := tx.Commit(); err != nil {
			return LogoutResult{}, fmt.Errorf("commit transaction: %w", err)
		}
		return LogoutResult{Revoked: true}, nil
	}

	if err := revokeSessionByID(ctx, tx, session.ID); err != nil {
		return LogoutResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return LogoutResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return LogoutResult{Revoked: true}, nil
}

type sessionRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt time.Time
}

func createSession(ctx context.Context, tx *sql.Tx, now time.Time, userID uuid.UUID, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) (AuthTokenPair, error) {
	refreshToken, err := generateToken(32)
	if err != nil {
		return AuthTokenPair{}, err
	}
	accessToken, err := generateToken(32)
	if err != nil {
		return AuthTokenPair{}, err
	}

	refreshExpiry := now.Add(refreshTokenTTL)
	refreshHash := hashRefreshToken(refreshToken)
	_, err = tx.ExecContext(ctx, `
INSERT INTO user_sessions (id, user_id, refresh_token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5)
`, uuid.New(), userID, refreshHash, refreshExpiry, now)
	if err != nil {
		return AuthTokenPair{}, fmt.Errorf("insert session: %w", err)
	}

	return AuthTokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		TokenType:    "bearer",
	}, nil
}

func emailExists(ctx context.Context, tx *sql.Tx, email string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func loadUserCredentials(ctx context.Context, tx *sql.Tx, email string) (uuid.UUID, string, error) {
	var userID uuid.UUID
	var passwordHash string
	err := tx.QueryRowContext(ctx, `SELECT id, password_hash FROM users WHERE email = $1`, email).Scan(&userID, &passwordHash)
	return userID, passwordHash, err
}

func loadSessionByHash(ctx context.Context, tx *sql.Tx, hash string) (sessionRecord, error) {
	var session sessionRecord
	var revokedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT id, user_id, expires_at, revoked_at
FROM user_sessions
WHERE refresh_token_hash = $1
`, hash).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &revokedAt)
	if err != nil {
		return sessionRecord{}, err
	}
	if revokedAt.Valid {
		session.RevokedAt = revokedAt.Time
	}
	return session, nil
}

func loadSessionByHashAllowRevoked(ctx context.Context, tx *sql.Tx, hash string) (sessionRecord, error) {
	var session sessionRecord
	var revokedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT id, user_id, expires_at, revoked_at
FROM user_sessions
WHERE refresh_token_hash = $1
`, hash).Scan(&session.ID, &session.UserID, &session.ExpiresAt, &revokedAt)
	if err != nil {
		return sessionRecord{}, err
	}
	if revokedAt.Valid {
		session.RevokedAt = revokedAt.Time
	}
	return session, nil
}

func lookupRevokedSession(ctx context.Context, tx *sql.Tx, hash string) (uuid.UUID, bool) {
	var userID uuid.UUID
	var revokedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
SELECT user_id, revoked_at
FROM user_sessions
WHERE refresh_token_hash = $1
`, hash).Scan(&userID, &revokedAt)
	if err != nil || !revokedAt.Valid {
		return uuid.UUID{}, false
	}
	return userID, true
}

func revokeSessionByID(ctx context.Context, tx *sql.Tx, sessionID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
UPDATE user_sessions
SET revoked_at = COALESCE(revoked_at, $2)
WHERE id = $1
`, sessionID, time.Now())
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func revokeAllUserSessions(ctx context.Context, tx *sql.Tx, userID uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `
UPDATE user_sessions
SET revoked_at = COALESCE(revoked_at, now())
WHERE user_id = $1
`, userID); err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegisterRequest(request RegisterRequest) *ValidationError {
	fields := make(map[string]string)
	if !looksLikeEmail(request.Email) {
		fields["email"] = "invalid_format"
	}
	if len(strings.TrimSpace(request.Password)) < 8 {
		fields["password"] = "too_short"
	}
	if len(strings.TrimSpace(request.FullName)) < 2 {
		fields["full_name"] = "too_short"
	}
	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Code: "validation_error", Message: "Invalid input", Fields: fields}
}

func validateLoginRequest(request LoginRequest) *ValidationError {
	fields := make(map[string]string)
	if !looksLikeEmail(request.Email) {
		fields["email"] = "invalid_format"
	}
	if len(strings.TrimSpace(request.Password)) < 8 {
		fields["password"] = "too_short"
	}
	if len(fields) == 0 {
		return nil
	}
	return &ValidationError{Code: "validation_error", Message: "Invalid input", Fields: fields}
}

func validateRefreshRequest(request RefreshRequest) *ValidationError {
	if strings.TrimSpace(request.RefreshToken) == "" {
		return &ValidationError{Code: "validation_error", Message: "Invalid input", Fields: map[string]string{"refresh_token": "required"}}
	}
	return nil
}

func validateVerifyEmailRequest(request VerifyEmailRequest) *ValidationError {
	if strings.TrimSpace(request.VerificationToken) == "" {
		return &ValidationError{Code: "validation_error", Message: "Invalid input", Fields: map[string]string{"verification_token": "required"}}
	}
	return nil
}

func looksLikeEmail(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.Contains(trimmed, " ") {
		return false
	}
	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 {
		return false
	}
	return parts[0] != "" && parts[1] != ""
}

func generateToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func hashVerificationToken(token string) string {
	return hashRefreshToken(token)
}
