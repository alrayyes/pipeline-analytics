package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"
)

// QuarantineDuration is how long a quarantine stays in force from the time it
// was marked or last renewed.
const QuarantineDuration = 30 * 24 * time.Hour

// QuarantineKey names the step a quarantine marks: one step name within one
// pipeline, the same identity FlakyStep uses.
type QuarantineKey struct {
	Pipeline PipelineRef
	Step     string
}

// Quarantine is a person's mark on a flaky step. It mutes the step's
// contribution to its pipeline's health signal and to the flaky-step ratio and
// changes no other figure.
type Quarantine struct {
	Note          string
	QuarantinedAt time.Time
	ExpiresAt     time.Time
}

// countsAsFlaky is the one place that decides whether a flaky step feeds a
// health signal or the flaky-step ratio: a quarantined step is still flaky, it
// just doesn't count.
func countsAsFlaky(step Step) bool {
	return step.Flaky && step.Quarantine == nil
}

// withQuarantines attaches the active mark to each step it names in pipeline.
func withQuarantines(steps []Step, pipeline PipelineRef, active map[QuarantineKey]Quarantine) []Step {
	for i := range steps {
		if q, ok := active[QuarantineKey{Pipeline: pipeline, Step: steps[i].Name}]; ok {
			steps[i].Quarantine = &q
		}
	}

	return steps
}

// MaxQuarantineNote is the longest note a quarantine takes, in characters.
const MaxQuarantineNote = 500

// ErrQuarantineNoteTooLong is returned for a note over MaxQuarantineNote
// characters.
var ErrQuarantineNoteTooLong = errors.New("quarantine note too long")

// QuarantineStep marks a step of a pipeline quarantined from now for
// QuarantineDuration, replacing any earlier mark. It does not check that the
// step is flaky: a mark on a step that isn't does nothing until it is. An
// unknown pipeline is ErrPipelineNotFound.
func (s *Service) QuarantineStep(ctx context.Context, id PipelineID, step, note string, now time.Time) (Quarantine, error) {
	if utf8.RuneCountInString(note) > MaxQuarantineNote {
		return Quarantine{}, ErrQuarantineNoteTooLong
	}

	ref, err := s.knownPipeline(ctx, id)
	if err != nil {
		return Quarantine{}, err
	}

	q, err := s.store.QuarantineStep(ctx, QuarantineKey{Pipeline: ref, Step: step}, note, now)
	if err != nil {
		return Quarantine{}, fmt.Errorf("quarantine step: %w", err)
	}

	return q, nil
}

// UnquarantineStep removes a step's mark. A step with none is fine; an unknown
// pipeline is ErrPipelineNotFound.
func (s *Service) UnquarantineStep(ctx context.Context, id PipelineID, step string) error {
	ref, err := s.knownPipeline(ctx, id)
	if err != nil {
		return err
	}

	if err := s.store.UnquarantineStep(ctx, QuarantineKey{Pipeline: ref, Step: step}); err != nil {
		return fmt.Errorf("un-quarantine step: %w", err)
	}

	return nil
}

// knownPipeline parses id and checks the pipeline has recorded runs.
func (s *Service) knownPipeline(ctx context.Context, id PipelineID) (PipelineRef, error) {
	ref, err := ParsePipelineID(id)
	if err != nil {
		return PipelineRef{}, err
	}

	runs, err := s.store.PipelineRuns(ctx, ref, Window{RunCount: 1})
	if err != nil {
		return PipelineRef{}, fmt.Errorf("load pipeline runs: %w", err)
	}

	if len(runs) == 0 {
		return PipelineRef{}, ErrPipelineNotFound
	}

	return ref, nil
}
