package actionlint

import (
	"strings"
	"testing"
)

// TestWorkflowKeyExprPolicyDoesNotOverlapWithAvailability checks that no key in the handwritten
// policy table is also defined in the generated availability table. When the official documentation
// starts to document context availability for one of these keys, the generated availability.go will
// contain the key too (after regenerating it), and the two tables would overlap. In that case the
// key must be removed from the handwritten table.
func TestWorkflowKeyExprPolicyDoesNotOverlapWithAvailability(t *testing.T) {
	for key := range workflowKeyExprPolicies {
		ctx, sp := WorkflowKeyAvailability(key)
		if ctx != nil || sp != nil {
			t.Errorf(
				"workflow key %q is defined in the generated availability table (contexts: %v, special functions: %v) but also listed in the handwritten policy table (workflow_key_policy.go). Remove the key from the handwritten table since the official documentation now documents its availability",
				key,
				ctx,
				sp,
			)
		}
	}
}

// TestWorkflowKeyExprPolicyEntries checks that each entry in the handwritten policy table is
// properly populated so that the classification can be reviewed against the official documentation
// later.
func TestWorkflowKeyExprPolicyEntries(t *testing.T) {
	if len(workflowKeyExprPolicies) == 0 {
		t.Fatal("the handwritten policy table is empty")
	}

	for key, p := range workflowKeyExprPolicies {
		t.Run(key, func(t *testing.T) {
			switch p.Policy {
			case WorkflowKeyNotEvaluated, WorkflowKeyTentative, WorkflowKeyAnyContext:
				// ok
			default:
				t.Errorf("unknown policy %d", p.Policy)
			}
			if strings.TrimSpace(p.Rationale) == "" {
				t.Error("rationale must not be empty")
			}
			if !strings.HasPrefix(p.DocURL, "https://docs.github.com/") {
				t.Errorf("documentation URL %q should start with https://docs.github.com/", p.DocURL)
			}

			gotPolicy, rationale, docURL, ok := WorkflowKeyExpressionPolicy(key)
			if !ok {
				t.Fatal("WorkflowKeyExpressionPolicy returned ok=false for a key in the table")
			}
			if gotPolicy != p.Policy || rationale != p.Rationale || docURL != p.DocURL {
				t.Errorf(
					"WorkflowKeyExpressionPolicy(%q) = (%v, %q, %q) but the table has (%v, %q, %q)",
					key, gotPolicy, rationale, docURL, p.Policy, p.Rationale, p.DocURL,
				)
			}
		})
	}
}

func TestWorkflowKeyExpressionPolicyUnknownKey(t *testing.T) {
	if _, _, _, ok := WorkflowKeyExpressionPolicy("jobs.<job_id>.totally-unknown"); ok {
		t.Error("ok should be false for a key not in the handwritten table")
	}
	// A key defined in the generated availability table must not be in the handwritten table.
	if _, _, _, ok := WorkflowKeyExpressionPolicy("jobs.<job_id>.steps.if"); ok {
		t.Error("a key from the generated availability table must not be in the handwritten table")
	}
}

// TestWorkflowKeyExprPolicyNotEvaluatedDefault asserts the classification of the only key that is
// currently confirmed as "not evaluated" and enabled by default.
func TestWorkflowKeyExprPolicyNotEvaluatedDefault(t *testing.T) {
	policy, _, docURL, ok := WorkflowKeyExpressionPolicy("jobs.<job_id>.uses")
	if !ok {
		t.Fatal("jobs.<job_id>.uses must be in the handwritten policy table")
	}
	if policy != WorkflowKeyNotEvaluated {
		t.Errorf("jobs.<job_id>.uses must be classified as not-evaluated but got %v", policy)
	}
	if !strings.Contains(docURL, "#jobsjob_iduses") {
		t.Errorf("unexpected documentation URL %q", docURL)
	}
}
