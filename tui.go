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
		return
	}

	if !continueFlow {
		return
	}

	// Before advancing, log the user's input to preserve history
	w.log("✓ " + w.label + ": " + newValue)

	// Move to next step
	w.currentStepIdx++
	w.initCurrentStep()
}

// Loggable implementation
func (w *Wizard) SetLog(f func(message ...any)) {
	w.log = f
	w.hasRealLogger = true
}

// StreamingLoggable implementation
func (w *Wizard) AlwaysShowAllLogs() bool { return true }

// TabAware implementation
func (w *Wizard) OnTabActive() {
	// Let the wizard intercept TabAware so it can manually trigger
	// the OnShow hook of the current step when the tab becomes active.
	w.showCurrentStep()

	for _, item := range w.items {
		if aware, ok := item.(interface{ OnTabActive() }); ok {
			aware.OnTabActive()
		}
	}
}

// Cancelable implementation
func (w *Wizard) Cancel() {
	w.waitingForUser = false
	w.log("Wizard cancelled")
}
