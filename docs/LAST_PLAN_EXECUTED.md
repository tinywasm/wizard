---
PLAN: "feat: Step.Sensitive so a secret-collecting step never logs its raw value"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — masked steps

## Part of a multi-repo wave — depends on `webtyp/tui` publishing first

This is one piece of `KEYRING_DOTENV_MASTER_PLAN.md` (orchestrator:
`webtyp.com/app-releases`, `docs/KEYRING_DOTENV_MASTER_PLAN.md`).

**Do not dispatch this plan until `webtyp.com/tui` has published the
`Sensitive` interface** (its own `docs/PLAN.md` in the same wave). This plan
does not need to `go get` that new version — see "Why no new dependency"
below — but it exists to satisfy that interface, so verify on
`https://pkg.go.dev/webtyp.com/tui` (or `go list -m -versions
webtyp.com/tui`) that a version with `type Sensitive interface {
Sensitive() bool }` exists before starting.

## Why

`webtyp/app` needs an interactive wizard step that collects a secret (a
credential to store in `webtyp/keyring`) without ever showing it on screen
or writing it into the TUI's log — that log streams over SSE
(`webtyp/devtui`'s `GET /logs`, also read by `webtyp/app`'s MCP tool
`app_get_logs`), so anything logged in plaintext leaves the machine.

Today `Wizard.Change` (in [`tui.go`](tui.go)) always logs the raw input after
a step advances:

```go
w.log("✓ " + w.label + ": " + newValue)
```

A step built to collect a secret would leak it right there. This plan closes
that for every step, not just a future secret one — any step can opt in.

## Why no new dependency

`*Wizard` already satisfies `tui.HandlerInteractive` (`Name`/`Label`/`Value`/
`Change`/`WaitingForUser`, see the "Handler Interface" comment atop
[`tui.go`](tui.go)) **without ever importing `webtyp.com/tui`** —
Go interfaces are structural, so this repo just names the same methods. Adding
`Sensitive() bool` to `*Wizard` is the same trick: it makes `*Wizard` also
satisfy `tui.Sensitive` for any caller that imports that package (`devtui`
does), with **no new entry in `go.mod`** here.

## The change

### 1. `Step` gains a `Sensitive` flag — file **[`step.go`](step.go)**

```go
// Step represents a single interaction or execution unit in the wizard.
// It is designed to be used as a literal for easy instantiation.
type Step struct {
	LabelText string
	DefaultFn func(ctx *context.Context) string
	OnInputFn func(input string, ctx *context.Context) (continueFlow bool, err error)
	OnShowFn  func(log func(message ...any))
	// Sensitive marks this step's collected value as a secret: the wizard
	// never logs the raw input for this step, and any consumer that checks
	// Wizard.Sensitive() (e.g. webtyp/devtui, via tui.Sensitive) masks it
	// on screen while it is being typed.
	Sensitive bool
}
```

Add the getter next to the other `Step` methods, same file:

```go
// IsSensitive reports Sensitive. A method, not the field itself, because
// orchestratorStep needs it and a Go type cannot have a field and a method
// share one name.
func (s *Step) IsSensitive() bool { return s.Sensitive }
```

### 2. `orchestratorStep` gains the same method — file **[`wizard.go`](wizard.go)**

```go
// orchestratorStep is the internal interface used by the Wizard.
// Both the provided Step struct and any custom implementation can satisfy this.
type orchestratorStep interface {
	Label() string
	DefaultValue(ctx *context.Context) string
	OnInput(input string, ctx *context.Context) (bool, error)
	OnShow(log func(message ...any))
	IsSensitive() bool
}
```

Verified: `*Step` is the only type in this repo implementing
`orchestratorStep` (`grep -n "OnInput(input string" wizard_test.go` finds no
hand-written fixture) — no other type needs a matching method added.

### 3. `Wizard` exposes it and stops logging raw sensitive input — file **[`tui.go`](tui.go)**

```go
func (w *Wizard) Sensitive() bool {
	if w.currentStepIdx >= len(w.steps) {
		return false // wizard finished, no current step to be sensitive about
	}
	return w.steps[w.currentStepIdx].IsSensitive()
}
```

And in `Change`, replace the unconditional log line:

```go
	// Before advancing, log the user's input to preserve history
	w.log("✓ " + w.label + ": " + newValue)
```

with:

```go
	// Before advancing, log the user's input to preserve history — except the
	// raw value of a sensitive step, which must never reach the log (it
	// streams over SSE, see webtyp/devtui GET /logs).
	confirmed := newValue
	if step.IsSensitive() {
		confirmed = "••••••••"
	}
	w.log("✓ " + w.label + ": " + confirmed)
```

`step` is already the local variable bound at the top of `Change` (`step :=
w.steps[w.currentStepIdx]`) — no new lookup needed.

## Anti-footguns

- **Do not** add a `Sensitive` parameter to `wizard.New` or `Wizard` itself —
  sensitivity is per-step (a wizard can mix ordinary and secret steps), never
  a whole-wizard setting.
- **Do not** import `webtyp.com/tui` in this repo to "make it
  official" — see "Why no new dependency" above; the structural match is the
  existing, deliberate pattern in this package.
- `OnShowFn` is untouched — only the auto-log line in `Change` is masked.
  If a step's own `OnShowFn` or `OnInputFn` logs the value itself, that is
  outside this plan's control (`webtyp/app`'s secret step, in the plan that
  consumes this one, must not do that either — noted there, not here).

## Tests

File: **`sensitive_test.go`** (new, next to `wizard_test.go`). Build wizards
the same way `TestWizardFlow` does: a `MockModule{steps: []*Step{...}}`
(fixture already defined in `wizard_test.go`, same package) passed to
`New(onComplete, mod)`. None of the existing tests call `SetLog` — for the two
log-content tests, call `w.SetLog(func(msgs ...any) { ... capture into a
[]string ... })` directly (it is an exported method) before driving `Change`.

| Test | Asserts |
|---|---|
| `TestStepIsSensitiveDefaultFalse` | a zero-value `Step{}` → `IsSensitive()` is `false` |
| `TestStepIsSensitiveTrue` | `Step{Sensitive: true}` → `IsSensitive()` is `true` |
| `TestWizardSensitiveReflectsCurrentStep` | a `Wizard` built from two steps, first `Sensitive: false`, second `Sensitive: true` — `Wizard.Sensitive()` is `false` while on step 1, becomes `true` after advancing (`Change` with valid input) to step 2 |
| `TestWizardSensitiveFalseAfterCompletion` | a one-step wizard, `Sensitive: true` — after `Change` completes the only step (wizard finished, `currentStepIdx >= len(steps)`), `Wizard.Sensitive()` is `false` |
| `TestChangeMasksSensitiveStepInLog` | a wizard with one `Sensitive: true` step, `SetLog` capturing every message — after `Change("super-secret-value")`, no captured message contains the substring `"super-secret-value"` |
| `TestChangeLogsPlainStepNormally` | same setup with `Sensitive: false` — a captured message **does** contain the input value, unchanged behavior |

## Acceptance criteria

- [ ] `go build ./...` and `go vet ./...` clean.
- [ ] `go test ./...` green, including all new tests, and every pre-existing
      test in this repo still passes (no existing behavior changed for a
      non-sensitive step).
- [ ] `grep -n "IsSensitive() bool" step.go wizard.go tui.go` → three matches.
- [ ] `grep -n "newValue" tui.go` inside `Change` no longer appears directly
      inside the `w.log(...)` call — it is only used to build `confirmed`.

## Out of scope

The actual secret-collecting `Step` that calls into `webtyp/keyring` is
`webtyp/app`'s plan in this same wave, dispatched after this one and after
`webtyp/keyring` publish. Do not add any keyring-aware code here — this
repo has no dependency on `webtyp/keyring` and should not gain one.
