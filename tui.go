package wizard

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
	continueFlow, err := step.OnInput(newValue, w.ctx)

	if err != nil {
		w.log("Error: " + err.Error())
		// If error occurs, we stay on the current step unless we want specific behaviors.
		// For now, any error blocks progress.
		return
	}

	if !continueFlow {
		// Step logic decided not to proceed yet (e.g. waiting for more input)
		return
	}

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
