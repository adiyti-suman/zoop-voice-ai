package verification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"verification-platform/internal/domain/verification"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, v *verification.Verification) error {
	metaBytes, err := v.GetMetadataBytes()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO verifications (
			id, external_id, type, status, subject_id, workflow_id, workflow_version, 
			language, locale, metadata, created_at, updated_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`
	_, err = r.db.ExecContext(ctx, query,
		v.ID, v.ExternalID, v.Type, v.Status, v.SubjectID, v.WorkflowID, v.WorkflowVersion,
		v.Language, v.Locale, metaBytes, v.CreatedAt, v.UpdatedAt, v.ExpiresAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*verification.Verification, error) {
	query := `
		SELECT id, external_id, type, status, subject_id, workflow_id, workflow_version, 
		       language, locale, metadata, created_at, updated_at, started_at, completed_at, expires_at
		FROM verifications
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanVerification(row)
}

func (r *PostgresRepository) GetByExternalID(ctx context.Context, externalID string) (*verification.Verification, error) {
	query := `
		SELECT id, external_id, type, status, subject_id, workflow_id, workflow_version, 
		       language, locale, metadata, created_at, updated_at, started_at, completed_at, expires_at
		FROM verifications
		WHERE external_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, externalID)
	return r.scanVerification(row)
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, expectedStatus verification.VerificationStatus, newStatus verification.VerificationStatus) error {
	query := `
		UPDATE verifications
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND status = $3
	`
	result, err := r.db.ExecContext(ctx, query, newStatus, id, expectedStatus)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Could mean it doesn't exist, or status changed concurrently
		return verification.ErrInvalidStateTransition
	}
	return nil
}

func (r *PostgresRepository) Update(ctx context.Context, v *verification.Verification) error {
	metaBytes, err := v.GetMetadataBytes()
	if err != nil {
		return err
	}

	query := `
		UPDATE verifications
		SET external_id = $1, type = $2, status = $3, subject_id = $4, workflow_id = $5, 
		    workflow_version = $6, language = $7, locale = $8, metadata = $9, 
		    updated_at = NOW(), started_at = $10, completed_at = $11, expires_at = $12
		WHERE id = $13
	`
	result, err := r.db.ExecContext(ctx, query,
		v.ExternalID, v.Type, v.Status, v.SubjectID, v.WorkflowID,
		v.WorkflowVersion, v.Language, v.Locale, metaBytes,
		v.StartedAt, v.CompletedAt, v.ExpiresAt, v.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return verification.ErrVerificationNotFound
	}
	return nil
}

func (r *PostgresRepository) LogAuditEvent(ctx context.Context, event *verification.AuditEvent) error {
	metaBytes, err := jsonMarshal(event.Metadata)
	if err != nil {
		return err
	}
	if len(metaBytes) == 0 || string(metaBytes) == "null" {
		metaBytes = []byte("{}")
	}

	query := `
		INSERT INTO verification_audit_events (
			id, verification_id, event_type, actor_type, actor_id, 
			previous_state, new_state, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.VerificationID, event.EventType, event.ActorType, event.ActorID,
		event.PreviousState, event.NewState, metaBytes, event.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) scanVerification(row *sql.Row) (*verification.Verification, error) {
	var v verification.Verification
	var metaBytes []byte

	err := row.Scan(
		&v.ID, &v.ExternalID, &v.Type, &v.Status, &v.SubjectID, &v.WorkflowID, &v.WorkflowVersion,
		&v.Language, &v.Locale, &metaBytes, &v.CreatedAt, &v.UpdatedAt, &v.StartedAt, &v.CompletedAt, &v.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, verification.ErrVerificationNotFound
		}
		return nil, err
	}

	if err := v.SetMetadataBytes(metaBytes); err != nil {
		return nil, err
	}
	return &v, nil
}

// jsonMarshal handles marshaling a map, returning empty byte array on nil
func jsonMarshal(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
