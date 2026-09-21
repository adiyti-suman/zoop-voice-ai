package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	wf "verification-platform/internal/domain/workflow"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// ── Workflow ──────────────────────────────────────────────────────────────────

func (r *PostgresRepository) CreateWorkflow(ctx context.Context, w *wf.Workflow) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workflows (id, name, description, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		w.ID, w.Name, w.Description, w.Status, w.CreatedAt, w.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetWorkflow(ctx context.Context, id string) (*wf.Workflow, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, status, created_at, updated_at FROM workflows WHERE id = $1`, id)
	var w wf.Workflow
	if err := row.Scan(&w.ID, &w.Name, &w.Description, &w.Status, &w.CreatedAt, &w.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wf.ErrWorkflowNotFound
		}
		return nil, err
	}
	return &w, nil
}

// ── Version ───────────────────────────────────────────────────────────────────

func (r *PostgresRepository) CreateVersion(ctx context.Context, v *wf.WorkflowVersion) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workflow_versions (id, workflow_id, version, status, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		v.ID, v.WorkflowID, v.Version, v.Status, v.CreatedAt,
	)
	return err
}

func (r *PostgresRepository) GetVersion(ctx context.Context, id string) (*wf.WorkflowVersion, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workflow_id, version, status, created_at, published_at, archived_at
		 FROM workflow_versions WHERE id = $1`, id)
	var v wf.WorkflowVersion
	if err := row.Scan(&v.ID, &v.WorkflowID, &v.Version, &v.Status,
		&v.CreatedAt, &v.PublishedAt, &v.ArchivedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wf.ErrVersionNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *PostgresRepository) UpdateVersionStatus(ctx context.Context, id string, status wf.WorkflowVersionStatus) error {
	now := time.Now().UTC()
	var extra string
	var extraVal interface{}
	switch status {
	case wf.VersionStatusPublished:
		extra, extraVal = ", published_at = $3", now
	case wf.VersionStatusArchived:
		extra, extraVal = ", archived_at = $3", now
	}

	var err error
	if extraVal != nil {
		_, err = r.db.ExecContext(ctx,
			`UPDATE workflow_versions SET status = $1, `+extra[2:]+" WHERE id = $2",
			status, id, extraVal)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE workflow_versions SET status = $1 WHERE id = $2`, status, id)
	}
	return err
}

// ── Graph ─────────────────────────────────────────────────────────────────────

func (r *PostgresRepository) UpsertNodes(ctx context.Context, nodes []*wf.WorkflowNode) error {
	for _, n := range nodes {
		cfg, _ := n.ConfigBytes()
		pos, _ := n.PositionBytes()
		var rp []byte
		if n.RetryPolicy != nil {
			rp, _ = json.Marshal(n.RetryPolicy)
		}
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO workflow_nodes
			   (id, workflow_version_id, node_key, type, name, config, position, timeout_ms, retry_policy, created_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			 ON CONFLICT (workflow_version_id, node_key)
			 DO UPDATE SET type=$4, name=$5, config=$6, position=$7, timeout_ms=$8, retry_policy=$9`,
			n.ID, n.WorkflowVersionID, n.NodeKey, n.Type, n.Name, cfg, pos, n.TimeoutMs, rp, n.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) UpsertEdges(ctx context.Context, edges []*wf.WorkflowEdge) error {
	for _, e := range edges {
		meta, _ := e.MetadataBytes()
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO workflow_edges
			   (id, workflow_version_id, source_node_key, target_node_key, condition, metadata, created_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)
			 ON CONFLICT DO NOTHING`,
			e.ID, e.WorkflowVersionID, e.SourceNodeKey, e.TargetNodeKey, e.Condition, meta, e.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) GetNodes(ctx context.Context, versionID string) ([]*wf.WorkflowNode, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workflow_version_id, node_key, type, name, config, position, timeout_ms, retry_policy, created_at
		 FROM workflow_nodes WHERE workflow_version_id = $1`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nodes []*wf.WorkflowNode
	for rows.Next() {
		var n wf.WorkflowNode
		var cfg, pos []byte
		var rp []byte
		if err := rows.Scan(&n.ID, &n.WorkflowVersionID, &n.NodeKey, &n.Type, &n.Name,
			&cfg, &pos, &n.TimeoutMs, &rp, &n.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(cfg, &n.Config)
		json.Unmarshal(pos, &n.Position)
		if rp != nil {
			json.Unmarshal(rp, &n.RetryPolicy)
		}
		nodes = append(nodes, &n)
	}
	return nodes, rows.Err()
}

func (r *PostgresRepository) GetEdges(ctx context.Context, versionID string) ([]*wf.WorkflowEdge, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workflow_version_id, source_node_key, target_node_key, condition, metadata, created_at
		 FROM workflow_edges WHERE workflow_version_id = $1`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var edges []*wf.WorkflowEdge
	for rows.Next() {
		var e wf.WorkflowEdge
		var meta []byte
		if err := rows.Scan(&e.ID, &e.WorkflowVersionID, &e.SourceNodeKey, &e.TargetNodeKey,
			&e.Condition, &meta, &e.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(meta, &e.Metadata)
		edges = append(edges, &e)
	}
	return edges, rows.Err()
}

// ── Runs ──────────────────────────────────────────────────────────────────────

func (r *PostgresRepository) CreateRun(ctx context.Context, run *wf.WorkflowRun) error {
	inp, _ := run.InputBytes()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workflow_runs
		   (id, verification_id, workflow_version_id, status, input, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		run.ID, run.VerificationID, run.WorkflowVersionID, run.Status, inp, run.CreatedAt, run.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetRun(ctx context.Context, id string) (*wf.WorkflowRun, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, verification_id, workflow_version_id, status, current_node_key,
		        input, output, error, started_at, completed_at, created_at, updated_at
		 FROM workflow_runs WHERE id = $1`, id)
	var run wf.WorkflowRun
	var inp, out, errB []byte
	if err := row.Scan(&run.ID, &run.VerificationID, &run.WorkflowVersionID, &run.Status,
		&run.CurrentNodeKey, &inp, &out, &errB,
		&run.StartedAt, &run.CompletedAt, &run.CreatedAt, &run.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, wf.ErrRunNotFound
		}
		return nil, err
	}
	json.Unmarshal(inp, &run.Input)
	json.Unmarshal(out, &run.Output)
	if errB != nil && string(errB) != "null" {
		json.Unmarshal(errB, &run.Error)
	}
	return &run, nil
}

func (r *PostgresRepository) UpdateRunStatus(ctx context.Context, id string, status wf.WorkflowRunStatus, runErr *wf.RunError) error {
	now := time.Now().UTC()
	var errBytes []byte
	if runErr != nil {
		errBytes, _ = json.Marshal(runErr)
	}

	var completedAt *time.Time
	if wf.IsRunTerminal(status) {
		completedAt = &now
	}

	var startedAt *time.Time
	if status == wf.RunStatusRunning {
		startedAt = &now
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_runs
		 SET status=$1, error=$2, updated_at=$3, completed_at=COALESCE($4, completed_at),
		     started_at=COALESCE($5, started_at)
		 WHERE id=$6`,
		status, errBytes, now, completedAt, startedAt, id,
	)
	return err
}

func (r *PostgresRepository) UpdateRunCurrentNode(ctx context.Context, id string, nodeKey string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_runs SET current_node_key=$1, updated_at=$2 WHERE id=$3`,
		nodeKey, time.Now().UTC(), id,
	)
	return err
}

func (r *PostgresRepository) SetRunOutput(ctx context.Context, id string, output map[string]interface{}) error {
	b, _ := json.Marshal(output)
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_runs SET output=$1, updated_at=$2 WHERE id=$3`,
		b, time.Now().UTC(), id,
	)
	return err
}

// ── Node Runs ─────────────────────────────────────────────────────────────────

func (r *PostgresRepository) CreateNodeRun(ctx context.Context, nr *wf.NodeRun) error {
	inp, _ := nr.InputBytes()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO workflow_node_runs
		   (id, workflow_run_id, node_key, attempt, status, input, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		nr.ID, nr.WorkflowRunID, nr.NodeKey, nr.Attempt, nr.Status, inp, nr.CreatedAt, nr.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateNodeRun(ctx context.Context, nr *wf.NodeRun) error {
	out, _ := nr.OutputBytes()
	var errBytes []byte
	if nr.Error != nil {
		errBytes, _ = json.Marshal(nr.Error)
	}
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`UPDATE workflow_node_runs
		 SET status=$1, output=$2, error=$3, completed_at=$4, duration_ms=$5, updated_at=$6,
		     started_at=COALESCE($7, started_at)
		 WHERE id=$8`,
		nr.Status, out, errBytes, nr.CompletedAt, nr.DurationMs, now, nr.StartedAt, nr.ID,
	)
	return err
}

func (r *PostgresRepository) GetNodeRuns(ctx context.Context, runID string) ([]*wf.NodeRun, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workflow_run_id, node_key, attempt, status, input, output, error,
		        started_at, completed_at, duration_ms, created_at, updated_at
		 FROM workflow_node_runs WHERE workflow_run_id=$1 ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nrs []*wf.NodeRun
	for rows.Next() {
		var nr wf.NodeRun
		var inp, out, errB []byte
		if err := rows.Scan(&nr.ID, &nr.WorkflowRunID, &nr.NodeKey, &nr.Attempt, &nr.Status,
			&inp, &out, &errB, &nr.StartedAt, &nr.CompletedAt, &nr.DurationMs,
			&nr.CreatedAt, &nr.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(inp, &nr.Input)
		json.Unmarshal(out, &nr.Output)
		if errB != nil && string(errB) != "null" {
			json.Unmarshal(errB, &nr.Error)
		}
		nrs = append(nrs, &nr)
	}
	return nrs, rows.Err()
}

// ensure compile-time interface satisfaction
var _ Repository = (*PostgresRepository)(nil)

// newUUID is a helper to keep callers cleaner
func newUUID() uuid.UUID { return uuid.New() }
