package actionlint

// This file contains a handwritten table of policies for ${{ }} expressions at workflow keys which
// have no entry in the "Context availability" table of the official documentation. Such keys are
// not covered by the generated availability.go (see availability.go and
// scripts/generate-availability).
//
// The table must not overlap with the generated availability table. If the official document starts
// to document context availability for one of these keys, the generated table (availability.go) will
// contain it too, and TestWorkflowKeyExprPolicyDoesNotOverlapWithAvailability fails to make us
// notice. In such a case, remove the key from this handwritten table.

const workflowSyntaxDocURL = "https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax"

// WorkflowKeyExprPolicy is a policy for how actionlint treats ${{ }} expressions at a workflow key
// which is not covered by the generated context availability table.
type WorkflowKeyExprPolicy int

const (
	// WorkflowKeyNotEvaluated means the official documentation explicitly states that expressions
	// cannot be used at this key. A ${{ }} placeholder is treated as a literal string or rejected by
	// the platform, so it is never evaluated. The dedicated diagnostic is enabled by default.
	WorkflowKeyNotEvaluated WorkflowKeyExprPolicy = iota
	// WorkflowKeyTentative means it is not confirmed by the official documentation whether the
	// platform evaluates expressions at this key. The dedicated diagnostic is disabled by default and
	// can be enabled through the "unevaluated-expressions.tentative" configuration.
	WorkflowKeyTentative
	// WorkflowKeyAnyContext means expressions at this key are evaluated with any context available.
	// No dedicated diagnostic is reported.
	WorkflowKeyAnyContext
)

func (p WorkflowKeyExprPolicy) String() string {
	switch p {
	case WorkflowKeyNotEvaluated:
		return "not-evaluated"
	case WorkflowKeyTentative:
		return "tentative"
	case WorkflowKeyAnyContext:
		return "any-context"
	default:
		return "unknown"
	}
}

type workflowKeyExprPolicy struct {
	// Policy is the classification of the workflow key.
	Policy WorkflowKeyExprPolicy
	// Rationale describes the basis of the classification. It is used for documentation and tests
	// so that the classification can be reviewed against the official document later.
	Rationale string
	// DocURL is a link to the relevant section of the official documentation. It is appended to the
	// diagnostic message.
	DocURL string
}

// tentativeRationale is the common basis for the "tentative" classification.
const tentativeRationale = "the official documentation does not state whether expressions are supported at this key, and the key is not listed in the context availability table"

// workflowKeyExprPolicies is a handwritten table from a workflow key to its expression policy. See
// the comment at the top of this file for the constraints of this table.
var workflowKeyExprPolicies = map[string]workflowKeyExprPolicy{
	// Not evaluated. The official documentation explicitly says "You cannot use contexts or
	// expressions in this keyword." for the reusable workflow call by "uses:".
	"jobs.<job_id>.uses": {
		Policy:    WorkflowKeyNotEvaluated,
		Rationale: "the official documentation says \"You cannot use contexts or expressions in this keyword\" at jobs.<job_id>.uses",
		DocURL:    workflowSyntaxDocURL + "#jobsjob_iduses",
	},

	// Tentative. The official documentation does not confirm whether expressions are evaluated and
	// the key is not listed in the context availability table.
	"name": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#name",
	},
	"on.<event_name>.types": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onevent_nametypes",
	},
	"on.repository_dispatch.types": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onevent_nametypes",
	},
	"on.workflow_run.workflows": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onworkflow_runworkflows",
	},
	"on.<event>.<filter>": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onpushbranchestagsbranches-ignoretags-ignore",
	},
	"on.schedule.cron": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onschedule",
	},
	"on.schedule.timezone": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onschedule",
	},
	"on.workflow_dispatch.inputs.<input_id>.description": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onworkflow_dispatchinputsinput_iddescription",
	},
	"on.workflow_dispatch.inputs.<input_id>.default": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onworkflow_dispatchinputsinput_iddefault",
	},
	"on.workflow_dispatch.inputs.<input_id>.options": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onworkflow_dispatchinputsinput_idoptions",
	},
	"on.workflow_call.inputs.<input_id>.description": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onworkflow_callinputs",
	},
	"on.workflow_call.secrets.<secret_id>.description": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#onworkflow_callsecrets",
	},
	"jobs.<job_id>.needs": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#jobsjob_idneeds",
	},
	"jobs.<job_id>.steps.shell": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#jobsjob_idstepsshell",
	},
	"jobs.<job_id>.steps.uses": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#jobsjob_idstepsuses",
	},
	"jobs.<job_id>.steps.id": {
		Policy:    WorkflowKeyTentative,
		Rationale: tentativeRationale,
		DocURL:    workflowSyntaxDocURL + "#jobsjob_idstepsid",
	},
}

// WorkflowKeyExpressionPolicy returns the expression policy for the given workflow key. The second
// and third return values are the rationale and the documentation URL of the classification. When
// the key is not in the handwritten table, ok is false and the caller should fall back to the
// generated context availability logic.
func WorkflowKeyExpressionPolicy(key string) (policy WorkflowKeyExprPolicy, rationale, docURL string, ok bool) {
	p, ok := workflowKeyExprPolicies[key]
	if !ok {
		return 0, "", "", false
	}
	return p.Policy, p.Rationale, p.DocURL, true
}
