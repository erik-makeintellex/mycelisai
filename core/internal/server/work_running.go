package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/mycelis/core/pkg/protocol"
)

// The "what's running" view is a bounded, read-only projection over the
// durable team_work_items spine plus the OutcomeProject.work_item_refs link.
// It performs no writes, emits no events, touches no NATS or in-memory
// registry, and holds no state between requests.
const (
	workRunningDefaultLimit   = 50
	workRunningMaxLimit       = 200
	workRunningSourceDatabase = "database"
	workRunningSourceTeamWork = "team_work_items"
	workRunningSourceOutcomes = "outcome_projects"
	workRunningRecoveryHint   = "Durable work storage is unavailable, so running work cannot be shown. Check Core database health, then refresh."
)

const workRunningItemsQuery = `
		SELECT id::text, team_id, COALESCE(run_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), objective, scope, owner,
		       execution_shape, execution_mode, work_intent, expected_outputs, expected_proof, capability_requirements,
		       governance_posture, state, COALESCE(last_event, 'null'::jsonb), needs_operator,
		       degradation_state, recovery_options, output_refs, proof_refs, audit_refs,
		       created_at, updated_at, version
		FROM team_work_items
		WHERE tenant_id='default' AND state NOT IN ('archived','output_ready')
		ORDER BY updated_at DESC, id ASC
		LIMIT $1`

const workRunningOutcomesQuery = `
		SELECT id::text, outcome_id, title, purpose, execution_mode, workspace_folder,
		       status, COALESCE(run_id,''), COALESCE(intent_proof_id,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), work_item_refs,
		       output_refs, proof_refs, recovery_refs, retention_policy,
		       created_at, updated_at, version
		FROM outcome_projects
		WHERE tenant_id='default' AND status <> 'archived' AND work_item_refs ?| $1::text[]
		ORDER BY updated_at DESC, id ASC`

type workRunningSourceError struct {
	source string
	err    error
}

func (e *workRunningSourceError) Error() string { return e.source + ": " + e.err.Error() }

// GET /api/v1/work/running
func (s *AdminServer) HandleWorkRunning(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}
	limit, err := parseWorkRunningLimit(r.URL.Query())
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	db := s.getDB()
	if db == nil {
		respondWorkRunningUnavailable(w, workRunningSourceDatabase)
		return
	}
	view, err := buildWorkRunningView(r.Context(), db, limit, time.Now().UTC())
	if err != nil {
		source := workRunningSourceDatabase
		var sourceErr *workRunningSourceError
		if errors.As(err, &sourceErr) {
			source = sourceErr.source
		}
		log.Printf("[work/running] %s read failed: %v", source, err)
		respondWorkRunningUnavailable(w, source)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(view))
}

func parseWorkRunningLimit(query url.Values) (int, error) {
	if !query.Has("limit") {
		return workRunningDefaultLimit, nil
	}
	limit, err := strconv.Atoi(strings.TrimSpace(query.Get("limit")))
	if err != nil || limit < 1 || limit > workRunningMaxLimit {
		return 0, errors.New("limit must be an integer between 1 and 200")
	}
	return limit, nil
}

// respondWorkRunningUnavailable fails closed: a 503 naming the failed source
// with a recovery hint and no item data, never a partial list.
func respondWorkRunningUnavailable(w http.ResponseWriter, source string) {
	respondAPIJSON(w, http.StatusServiceUnavailable, protocol.APIResponse{
		OK:    false,
		Error: "Running work is unavailable: " + source + " could not be read",
		Data:  protocol.WorkRunningFailure{Source: source, RecoveryHint: workRunningRecoveryHint},
	})
}

func buildWorkRunningView(ctx context.Context, db *sql.DB, limit int, now time.Time) (protocol.WorkRunningView, error) {
	view := protocol.WorkRunningView{
		GeneratedAt: now,
		Outcomes:    []protocol.WorkRunningOutcome{},
		Unassigned:  []protocol.WorkRunningTeamGroup{},
	}
	items, truncated, err := listWorkRunningItemsDB(ctx, db, limit)
	if err != nil {
		return protocol.WorkRunningView{}, &workRunningSourceError{source: workRunningSourceTeamWork, err: err}
	}
	view.Truncated = truncated
	if len(items) == 0 {
		return view, nil
	}
	projects, err := listWorkRunningOutcomesDB(ctx, db, items)
	if err != nil {
		return protocol.WorkRunningView{}, &workRunningSourceError{source: workRunningSourceOutcomes, err: err}
	}
	groupWorkRunningView(&view, items, projects)
	return view, nil
}

// listWorkRunningItemsDB is one statement for every team: no per-team loop.
// It reads limit+1 rows so truncation is reported rather than hidden.
func listWorkRunningItemsDB(ctx context.Context, db *sql.DB, limit int) ([]protocol.TeamWorkItem, bool, error) {
	rows, err := db.QueryContext(ctx, workRunningItemsQuery, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	items := make([]protocol.TeamWorkItem, 0, limit)
	truncated := false
	for rows.Next() {
		if len(items) == limit {
			truncated = true
			break
		}
		item, scanErr := scanTeamWorkItem(rows)
		if scanErr != nil {
			return nil, false, scanErr
		}
		if protocol.IsWorkRunningVisible(item) {
			items = append(items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	return items, truncated, nil
}

func listWorkRunningOutcomesDB(ctx context.Context, db *sql.DB, items []protocol.TeamWorkItem) ([]protocol.OutcomeProject, error) {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.WorkItemID)
	}
	rows, err := db.QueryContext(ctx, workRunningOutcomesQuery, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]protocol.OutcomeProject, 0)
	for rows.Next() {
		project, scanErr := scanOutcomeProject(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

// groupWorkRunningView assigns each item to exactly one Outcome: the most
// recently updated project that references it (query order, id tiebreak).
// Items without a referencing project are grouped by team under unassigned.
func groupWorkRunningView(view *protocol.WorkRunningView, items []protocol.TeamWorkItem, projects []protocol.OutcomeProject) {
	owner := make(map[string]int, len(items))
	for idx, project := range projects {
		if project.Status == protocol.OutcomeProjectStatusArchived {
			continue // archived Outcomes never group running work (SQL filters too)
		}
		for _, ref := range project.WorkItemRefs {
			if _, claimed := owner[ref]; !claimed {
				owner[ref] = idx
			}
		}
	}
	outcomeIndex := map[int]int{}
	outcomeProject := []int{}
	teamIndex := map[string]int{}
	for _, item := range items {
		entry := workRunningItem(item)
		countWorkRunningHealth(&view.Summary, entry.OutcomeHealth)
		projectIdx, linked := owner[item.WorkItemID]
		if !linked {
			idx, seen := teamIndex[item.TeamID]
			if !seen {
				idx = len(view.Unassigned)
				teamIndex[item.TeamID] = idx
				view.Unassigned = append(view.Unassigned, protocol.WorkRunningTeamGroup{TeamID: item.TeamID})
			}
			view.Unassigned[idx].Items = append(view.Unassigned[idx].Items, entry)
			continue
		}
		project := projects[projectIdx]
		entry.ProjectID = project.ProjectID
		idx, seen := outcomeIndex[projectIdx]
		if !seen {
			idx = len(view.Outcomes)
			outcomeIndex[projectIdx] = idx
			outcomeProject = append(outcomeProject, projectIdx)
			view.Outcomes = append(view.Outcomes, protocol.WorkRunningOutcome{
				ProjectID: project.ProjectID,
				OutcomeID: project.OutcomeID,
				Title:     project.Title,
			})
		}
		view.Outcomes[idx].Items = append(view.Outcomes[idx].Items, entry)
	}
	for idx := range view.Outcomes {
		outcome := &view.Outcomes[idx]
		states := []protocol.OutcomeHealthState{}
		// A project's own Completed/Archived label never outranks live items.
		projectHealth := protocol.OutcomeHealthForProject(projects[outcomeProject[idx]])
		if projectHealth != protocol.OutcomeHealthCompleted && projectHealth != protocol.OutcomeHealthArchived {
			states = append(states, projectHealth)
		}
		for _, entry := range outcome.Items {
			states = append(states, entry.OutcomeHealth)
		}
		outcome.OutcomeHealth = protocol.AggregateOutcomeHealth(states...)
	}
}

func workRunningItem(item protocol.TeamWorkItem) protocol.WorkRunningItem {
	return protocol.WorkRunningItem{
		WorkItemID:    item.WorkItemID,
		TeamID:        item.TeamID,
		RunID:         item.RunID,
		Title:         truncateWorkRunningTitle(item.Objective),
		OutcomeHealth: protocol.OutcomeHealthForTeamWork(item),
		UpdatedAt:     item.UpdatedAt,
		NeedsOperator: item.NeedsOperator,
		Details: protocol.WorkRunningItemDetails{
			State:            item.State,
			ExecutionShape:   item.ExecutionShape,
			DegradationState: item.DegradationState,
		},
	}
}

func truncateWorkRunningTitle(objective string) string {
	runes := []rune(strings.TrimSpace(objective))
	if len(runes) <= protocol.WorkRunningTitleMaxRunes {
		return string(runes)
	}
	return string(runes[:protocol.WorkRunningTitleMaxRunes])
}

func countWorkRunningHealth(summary *protocol.WorkRunningSummary, health protocol.OutcomeHealthState) {
	switch health {
	case protocol.OutcomeHealthRunning:
		summary.Running++
	case protocol.OutcomeHealthWaiting:
		summary.Waiting++
	case protocol.OutcomeHealthBlocked:
		summary.Blocked++
	case protocol.OutcomeHealthDegraded:
		summary.Degraded++
	}
}
