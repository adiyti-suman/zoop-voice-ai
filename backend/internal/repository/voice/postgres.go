package voice

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "verification-platform/internal/domain/voice"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session *domain.VoiceSession) error {
	query := `
		INSERT INTO voice_sessions (
			id, verification_id, workflow_run_id, status, language, sample_rate, channels, encoding, started_at, ended_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.VerificationID,
		session.WorkflowRunID,
		session.Status,
		session.Language,
		session.SampleRate,
		session.Channels,
		session.Encoding,
		session.StartedAt,
		session.EndedAt,
		session.CreatedAt,
		session.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetSession(ctx context.Context, id string) (*domain.VoiceSession, error) {
	query := `
		SELECT id, verification_id, workflow_run_id, status, language, sample_rate, channels, encoding, started_at, ended_at, created_at, updated_at
		FROM voice_sessions
		WHERE id = $1
	`
	var s domain.VoiceSession
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.VerificationID,
		&s.WorkflowRunID,
		&s.Status,
		&s.Language,
		&s.SampleRate,
		&s.Channels,
		&s.Encoding,
		&s.StartedAt,
		&s.EndedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) UpdateSessionStatus(ctx context.Context, id string, status domain.SessionStatus) error {
	query := `
		UPDATE voice_sessions
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	res, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}
