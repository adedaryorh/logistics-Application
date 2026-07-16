package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/adedaryorh/logistics-platform/services/identity-service/internal/model"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open identity database: %w", err)
	}

	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)

	return &Store{db: db}, nil
}

func (s *Store) Ping(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping identity database: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close identity database: %w", err)
	}
	return nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) CreateUser(ctx context.Context, fullName, email, passwordHash, role string, outboxPayload any) (*model.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer tx.Rollback()

	user := &model.User{}
	query := `
		INSERT INTO identity_.users (full_name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, full_name, email, phone, password_hash, email_verified, phone_verified, role, status, created_at, updated_at, deleted_at
	`
	if err := tx.QueryRowContext(ctx, query, fullName, email, passwordHash, role).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		return nil, fmt.Errorf("insert identity user: %w", err)
	}

	payloadMap := map[string]any{}
	if outboxPayload != nil {
		if cast, ok := outboxPayload.(map[string]any); ok {
			for key, value := range cast {
				payloadMap[key] = value
			}
		} else {
			return nil, fmt.Errorf("marshal user outbox payload: unsupported payload type %T", outboxPayload)
		}
	}
	payloadMap["user_id"] = user.ID

	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("marshal user outbox payload: %w", err)
	}

	eventID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate outbox id: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO identity_.outbox (id, aggregate_id, event_type, payload) VALUES ($1, $2, $3, $4)`,
		eventID.String(),
		user.ID,
		"user.registered",
		payloadBytes,
	); err != nil {
		return nil, fmt.Errorf("insert identity outbox event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create user transaction: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, full_name, email, phone, password_hash, email_verified, phone_verified, role, status, created_at, updated_at, deleted_at
		FROM identity_.users
		WHERE email = $1 AND deleted_at IS NULL
	`
	if err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, full_name, email, phone, password_hash, email_verified, phone_verified, role, status, created_at, updated_at, deleted_at
		FROM identity_.users
		WHERE id = $1 AND deleted_at IS NULL
	`
	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return user, nil
}

func (s *Store) UpsertOAuthUser(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin oauth upsert transaction: %w", err)
	}
	defer tx.Rollback()

	user := &model.User{}
	if err := tx.QueryRowContext(ctx, `
		SELECT u.id, u.full_name, u.email, u.phone, u.password_hash, u.email_verified, u.phone_verified, u.role, u.status, u.created_at, u.updated_at, u.deleted_at
		FROM identity_.oauth_accounts oa
		JOIN identity_.users u ON u.id = oa.user_id
		WHERE oa.provider = $1 AND oa.provider_user_id = $2 AND u.deleted_at IS NULL
	`, provider, providerUserID).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err == nil {
		if _, err := tx.ExecContext(ctx, `UPDATE identity_.users SET email_verified = email_verified OR $2, updated_at = NOW() WHERE id = $1`, user.ID, emailVerified); err != nil {
			return nil, fmt.Errorf("refresh oauth user verification status: %w", err)
		}
		if emailVerified {
			user.EmailVerified = true
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit oauth existing user transaction: %w", err)
		}
		return user, nil
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("query oauth account: %w", err)
	}

	if err := tx.QueryRowContext(ctx, `
		SELECT id, full_name, email, phone, password_hash, email_verified, phone_verified, role, status, created_at, updated_at, deleted_at
		FROM identity_.users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("query oauth user by email: %w", err)
		}

		insertQuery := `
			INSERT INTO identity_.users (full_name, email, password_hash, email_verified, role)
			VALUES (split_part($1, '@', 1), $1, '', $2, 'customer')
			RETURNING id, full_name, email, phone, password_hash, email_verified, phone_verified, role, status, created_at, updated_at, deleted_at
		`
		if err := tx.QueryRowContext(ctx, insertQuery, email, emailVerified).Scan(
			&user.ID,
			&user.FullName,
			&user.Email,
			&user.Phone,
			&user.PasswordHash,
			&user.EmailVerified,
			&user.PhoneVerified,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("insert oauth user: %w", err)
		}
	} else if emailVerified && !user.EmailVerified {
		if _, err := tx.ExecContext(ctx, `UPDATE identity_.users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`, user.ID); err != nil {
			return nil, fmt.Errorf("mark existing oauth email verified: %w", err)
		}
		user.EmailVerified = true
	}

	oauthID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate oauth account id: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity_.oauth_accounts (id, user_id, provider, provider_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, provider_user_id) DO NOTHING
	`, oauthID.String(), user.ID, provider, providerUserID); err != nil {
		return nil, fmt.Errorf("upsert oauth account: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit oauth upsert transaction: %w", err)
	}

	return user, nil
}

func (s *Store) CreateOAuthSession(ctx context.Context, session model.OAuthSession) error {
	query := `
		INSERT INTO identity_.oauth_sessions (
			id, provider, state, nonce, redirect_uri, code_challenge, code_challenge_method, expires_at, used_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if _, err := s.db.ExecContext(ctx, query,
		session.ID,
		session.Provider,
		session.State,
		session.Nonce,
		session.RedirectURI,
		session.CodeChallenge,
		session.CodeChallengeMethod,
		session.ExpiresAt,
		session.UsedAt,
		session.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert oauth session: %w", err)
	}
	return nil
}

func (s *Store) GetOAuthSessionByID(ctx context.Context, id string) (*model.OAuthSession, error) {
	session := &model.OAuthSession{}
	query := `
		SELECT id, provider, state, nonce, redirect_uri, code_challenge, code_challenge_method, expires_at, used_at, created_at
		FROM identity_.oauth_sessions
		WHERE id = $1
	`
	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.Provider,
		&session.State,
		&session.Nonce,
		&session.RedirectURI,
		&session.CodeChallenge,
		&session.CodeChallengeMethod,
		&session.ExpiresAt,
		&session.UsedAt,
		&session.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query oauth session by id: %w", err)
	}
	return session, nil
}

func (s *Store) UseOAuthSession(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.oauth_sessions SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id); err != nil {
		return fmt.Errorf("mark oauth session used: %w", err)
	}
	return nil
}

func (s *Store) DeleteExpiredOAuthSessions(ctx context.Context, before time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM identity_.oauth_sessions WHERE expires_at < $1 OR used_at IS NOT NULL`, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired oauth sessions: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("oauth session cleanup rows affected: %w", err)
	}
	return rows, nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, id string, phone *string) (*model.User, error) {
	user := &model.User{}
	query := `
		UPDATE identity_.users
		SET phone = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, full_name, email, phone, password_hash, email_verified, phone_verified, role, status, created_at, updated_at, deleted_at
	`
	if err := s.db.QueryRowContext(ctx, query, id, phone).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update user profile: %w", err)
	}
	return user, nil
}

func (s *Store) SoftDeleteUser(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.users SET status = 'deleted', deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return nil
}

func (s *Store) CreateRefreshToken(ctx context.Context, token model.RefreshToken) error {
	query := `
		INSERT INTO identity_.refresh_tokens (id, user_id, token_hash, family_id, device_id, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := s.db.ExecContext(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.FamilyID,
		token.DeviceID,
		token.ExpiresAt,
		token.RevokedAt,
		token.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (s *Store) GetRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	token := &model.RefreshToken{}
	query := `
		SELECT id, user_id, token_hash, family_id, device_id, expires_at, revoked_at, created_at
		FROM identity_.refresh_tokens
		WHERE token_hash = $1
	`
	if err := s.db.QueryRowContext(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.FamilyID,
		&token.DeviceID,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query refresh token by hash: %w", err)
	}
	return token, nil
}

func (s *Store) RevokeRefreshToken(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.refresh_tokens SET revoked_at = NOW() WHERE id = $1`, id); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (s *Store) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.refresh_tokens SET revoked_at = NOW() WHERE family_id = $1 AND revoked_at IS NULL`, familyID); err != nil {
		return fmt.Errorf("revoke refresh token family: %w", err)
	}
	return nil
}

func (s *Store) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens for user: %w", err)
	}
	return nil
}

func (s *Store) CreateAuditLog(ctx context.Context, userID *string, action, ip, userAgent string, metadata any) error {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate audit log id: %w", err)
	}

	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO identity_.audit_log (id, user_id, action, ip, user_agent, metadata) VALUES ($1, $2, $3, $4, $5, $6)`,
		id.String(),
		userID,
		action,
		ip,
		userAgent,
		metadataBytes,
	); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

func (s *Store) CreateMagicLink(ctx context.Context, magicLink model.MagicLink) error {
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO identity_.magic_links (id, user_id, token_hash, expires_at, used_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		magicLink.ID,
		magicLink.UserID,
		magicLink.TokenHash,
		magicLink.ExpiresAt,
		magicLink.UsedAt,
		magicLink.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert magic link: %w", err)
	}
	return nil
}

func (s *Store) CountRecentMagicLinksByUser(ctx context.Context, userID string, since time.Time) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM identity_.magic_links WHERE user_id = $1 AND created_at >= $2`, userID, since).Scan(&count); err != nil {
		return 0, fmt.Errorf("count recent magic links: %w", err)
	}
	return count, nil
}

func (s *Store) GetMagicLinkByHash(ctx context.Context, hash string) (*model.MagicLink, error) {
	link := &model.MagicLink{}
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at FROM identity_.magic_links WHERE token_hash = $1`
	if err := s.db.QueryRowContext(ctx, query, hash).Scan(&link.ID, &link.UserID, &link.TokenHash, &link.ExpiresAt, &link.UsedAt, &link.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query magic link by hash: %w", err)
	}
	return link, nil
}

func (s *Store) UseMagicLink(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.magic_links SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id); err != nil {
		return fmt.Errorf("mark magic link used: %w", err)
	}
	return nil
}

func (s *Store) MarkUserEmailVerified(ctx context.Context, userID string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("mark user email verified: %w", err)
	}
	return nil
}

func (s *Store) CreatePasswordReset(ctx context.Context, reset model.PasswordReset) error {
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO identity_.password_resets (id, user_id, token_hash, expires_at, used_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		reset.ID, reset.UserID, reset.TokenHash, reset.ExpiresAt, reset.UsedAt, reset.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert password reset: %w", err)
	}
	return nil
}

func (s *Store) GetPasswordResetByHash(ctx context.Context, hash string) (*model.PasswordReset, error) {
	reset := &model.PasswordReset{}
	query := `SELECT id, user_id, token_hash, expires_at, used_at, created_at FROM identity_.password_resets WHERE token_hash = $1`
	if err := s.db.QueryRowContext(ctx, query, hash).Scan(&reset.ID, &reset.UserID, &reset.TokenHash, &reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query password reset by hash: %w", err)
	}
	return reset, nil
}

func (s *Store) UsePasswordReset(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.password_resets SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`, id); err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	return nil
}

func (s *Store) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE identity_.users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, userID, passwordHash); err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	return nil
}
