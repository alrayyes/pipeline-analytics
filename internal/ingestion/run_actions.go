package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	// ErrRunNotFound is returned when a run id names no recorded run.
	ErrRunNotFound = errors.New("run not found")
	// ErrNotActionable means the run's state doesn't allow the action.
	ErrNotActionable = errors.New("run state doesn't allow this action")
	// ErrActionUnsupported means the forge has no API for the action.
	ErrActionUnsupported = errors.New("forge has no API for this action")
	// ErrActionForbidden means the stored token can't do the action.
	ErrActionForbidden = errors.New("token can't do this action")
	// ErrActionUnreachable means the forge didn't answer.
	ErrActionUnreachable = errors.New("forge unreachable")
)

// RunAction is something the dashboard can ask a forge to do to a run.
type RunAction string

// The actions on a run.
const (
	ActionRerun  RunAction = "rerun"
	ActionCancel RunAction = "cancel"
)

// ActionsFor lists what a session may ask the forge to do to a run in the
// given status: a re-run once it has concluded, a cancel before. Forgejo
// has no API for either, so it gets nothing. It's the one rule behind the
// run list's `actions` and the state check of the action endpoints.
func ActionsFor(forge Forge, status string) []RunAction {
	if forge != ForgeGitHub {
		return []RunAction{}
	}

	if status == runCompleted {
		return []RunAction{ActionRerun}
	}

	return []RunAction{ActionCancel}
}

const runCompleted = "completed"

// RunActionRequest names one run on its forge and what to do to it.
type RunActionRequest struct {
	// InstanceURL is set for Forgejo, empty for GitHub.
	InstanceURL string
	Identifier  string
	Token       string
	ForgeRunID  string
	// FailedOnly re-runs only the failed jobs. Unused by a cancel.
	FailedOnly bool
}

// RunActor asks a forge to re-run or cancel a run. It returns one of the
// ErrAction* errors, wrapped, when the forge refuses or can't be reached.
type RunActor interface {
	RerunRun(ctx context.Context, req RunActionRequest) error
	CancelRun(ctx context.Context, req RunActionRequest) error
}

// RunLookup finds a recorded run by its id, or ErrRunNotFound.
type RunLookup interface {
	RunByID(ctx context.Context, id string) (Run, error)
}

// RunActionRepos is the slice of the repo store the action service needs.
type RunActionRepos interface {
	GetRepo(ctx context.Context, id string) (Repo, error)
	RepoToken(ctx context.Context, id string) (string, error)
}

// RunActionService re-runs and cancels runs on their forge, using the
// repo's stored token (openspec/changes/forge-run-actions).
type RunActionService struct {
	runs   RunLookup
	repos  RunActionRepos
	actors map[Forge]RunActor
}

// NewRunActionService returns a RunActionService. A forge with no actor in
// actors answers ErrActionUnsupported.
func NewRunActionService(runs RunLookup, repos RunActionRepos, actors map[Forge]RunActor) *RunActionService {
	return &RunActionService{runs: runs, repos: repos, actors: actors}
}

// Rerun asks the forge to re-run a concluded run: its failed jobs when it
// failed, the whole run otherwise.
func (s *RunActionService) Rerun(ctx context.Context, runID string) error {
	return s.act(ctx, runID, ActionRerun)
}

// Cancel asks the forge to cancel a run that hasn't concluded.
func (s *RunActionService) Cancel(ctx context.Context, runID string) error {
	return s.act(ctx, runID, ActionCancel)
}

func (s *RunActionService) act(ctx context.Context, runID string, action RunAction) (err error) {
	run, err := s.runs.RunByID(ctx, runID)
	if err != nil {
		return fmt.Errorf("find run: %w", err)
	}

	defer func() {
		// Never the token: the run, the action and the outcome.
		slog.InfoContext(ctx, "run action", slog.String("run", runID), slog.String("action", string(action)), slog.String("outcome", outcome(err)))
	}()

	if (run.Status == runCompleted) != (action == ActionRerun) {
		return ErrNotActionable
	}

	repo, err := s.repos.GetRepo(ctx, run.RepoID)
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}

	actor, ok := s.actors[repo.Forge]
	if !ok {
		return ErrActionUnsupported
	}

	token, err := s.repos.RepoToken(ctx, repo.ID)
	if err != nil {
		return fmt.Errorf("get repo token: %w", err)
	}

	req := RunActionRequest{
		InstanceURL: repo.ForgejoInstanceURL,
		Identifier:  repo.Identifier,
		Token:       token,
		ForgeRunID:  run.ForgeRunID,
		FailedOnly:  run.Conclusion == "failure" || run.Conclusion == "timed_out",
	}

	if action == ActionCancel {
		err = actor.CancelRun(ctx, req)
	} else {
		err = actor.RerunRun(ctx, req)
	}

	if err != nil {
		return fmt.Errorf("%s run: %w", action, err)
	}

	return nil
}

func outcome(err error) string {
	switch {
	case err == nil:
		return "accepted"
	case errors.Is(err, ErrNotActionable):
		return "not_actionable"
	case errors.Is(err, ErrActionForbidden):
		return "forbidden"
	case errors.Is(err, ErrActionUnsupported):
		return "unsupported"
	case errors.Is(err, ErrActionUnreachable):
		return "unreachable"
	default:
		return "error"
	}
}
