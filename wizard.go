package wizard

import (
	"github.com/tinywasm/context"
	"github.com/tinywasm/fmt"
)

// Step represents a single interaction or execution unit in the wizard
type Step interface {
	// Label returns the prompt text for the UI (e.g., "Project Name")
	Label() string

	// DefaultValue returns a suggestion based on the current context
	DefaultValue(ctx *context.Context) string

	// OnInput executes the step logic.
	OnInput(input string, ctx *context.Context) (newCtx *context.Context, continueFlow bool, err error)
}

// Module represents a pluggable component that provides a sequence of steps
type Module interface {
	// Name returns the module identifier
	Name() string

	// GetSteps returns the sequence of steps this module requires.
	GetSteps() []any
}

// Wizard orchestrates the execution of multiple Steps
type Wizard struct {
	log func(message ...any)

	// Orchestration state
	ctx            *context.Context
	steps          []Step
	currentStepIdx int

	// TUI state
	stepMessage    string
	label          string
	currentValue   string
	waitingForUser bool

	onComplete func()
}

// New creates a generic wizard with the given steps
func New(onComplete func(), steps ...Step) *Wizard {
	w := &Wizard{
		log:            func(...any) {},
		ctx:            context.Background(),
		steps:          steps,
		currentStepIdx: 0,
		onComplete:     onComplete,
		stepMessage:    "WIZARD",
	}
	w.initCurrentStep()
	return w
}

func (w *Wizard) initCurrentStep() {
	if w.currentStepIdx >= len(w.steps) {
		// All steps done
		w.waitingForUser = false
		w.label = "Complete"
		w.stepMessage = "DONE"
		w.currentValue = "Wizard completed successfully."

		if w.onComplete != nil {
			w.onComplete()
			w.onComplete = nil
		}
		return
	}

	step := w.steps[w.currentStepIdx]
	w.label = step.Label()
	w.currentValue = step.DefaultValue(w.ctx)
	w.stepMessage = "STEP " + fmt.Convert(w.label).PathBase().String() // Simple header
	w.waitingForUser = true
}

// Handler Interface (compatible with app's TUI expectations)

func (w *Wizard) Name() string         { return w.stepMessage }
func (w *Wizard) Label() string        { return w.label }
func (w *Wizard) Value() string        { return w.currentValue }
func (w *Wizard) WaitingForUser() bool { return w.waitingForUser }

func (w *Wizard) Change(newValue string) {
	if w.currentStepIdx >= len(w.steps) {
		return
	}

	step := w.steps[w.currentStepIdx]
	newCtx, continueFlow, err := step.OnInput(newValue, w.ctx)

	if err != nil {
		w.log("Error: " + err.Error())
		if !continueFlow {
			// If cannot continue, stay on current step (user must retry)
			return
		}
		// If continueFlow is true, we log error but proceed
	}

	w.ctx = newCtx

	// Move to next step
	w.currentStepIdx++
	w.initCurrentStep()
}

// Loggable implementation
func (w *Wizard) SetLog(f func(message ...any)) { w.log = f }

// StreamingLoggable implementation
func (w *Wizard) AlwaysShowAllLogs() bool { return true }

// Cancelable implementation
func (w *Wizard) Cancel() {
	w.waitingForUser = false
	w.log("Wizard cancelled")
}
