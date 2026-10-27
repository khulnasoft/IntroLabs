package generator

// TemplateState tracks the lifecycle state of a generated lab template.
type TemplateState string

const (
	StateDraft             TemplateState = "draft"
	StateMetadataOnly      TemplateState = "metadata-only"
	StateNeedsVerification TemplateState = "needs-verification"
	StateVerified          TemplateState = "verified"
	StatePublished         TemplateState = "published"
	StateDeprecated        TemplateState = "deprecated"
)

// ValidStates lists all valid template states in order.
var ValidStates = []TemplateState{
	StateDraft,
	StateMetadataOnly,
	StateNeedsVerification,
	StateVerified,
	StatePublished,
	StateDeprecated,
}

// IsValid reports whether s is a recognized template state.
func (s TemplateState) IsValid() bool {
	for _, st := range ValidStates {
		if s == st {
			return true
		}
	}
	return false
}

// CanTransition reports whether a transition from fromState to toState is allowed.
// Transitions must proceed forward through the state machine; skipping states is not allowed.
func CanTransition(fromState, toState TemplateState) bool {
	if !fromState.IsValid() || !toState.IsValid() {
		return false
	}
	fromIdx := -1
	toIdx := -1
	for i, st := range ValidStates {
		if st == fromState {
			fromIdx = i
		}
		if st == toState {
			toIdx = i
		}
	}
	if fromIdx == -1 || toIdx == -1 {
		return false
	}
	// Allow staying in the same state (idempotent) or advancing by exactly one.
	return toIdx == fromIdx || toIdx == fromIdx+1
}

// NextState returns the next state in the lifecycle, or the current state if already at the end.
func NextState(s TemplateState) TemplateState {
	for i, st := range ValidStates {
		if st == s && i+1 < len(ValidStates) {
			return ValidStates[i+1]
		}
	}
	return s
}

// IsPublishable reports whether a template in this state can be published.
func (s TemplateState) IsPublishable() bool {
	return s == StateVerified
}

// IsRunnable reports whether a template in this state can be launched.
func (s TemplateState) IsRunnable() bool {
	return s == StateVerified || s == StatePublished
}

// String returns the state as a string.
func (s TemplateState) String() string {
	return string(s)
}
