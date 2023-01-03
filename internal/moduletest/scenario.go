package moduletest

import (
	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/checks"
	"github.com/hashicorp/terraform/internal/states"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

// ScenarioResult represents the overall results of executing a single test
// scenario.
type ScenarioResult struct {
	// Name is the user-selected name for the scenario.
	Name string

	// Status is the aggregate status across all of the steps. This uses the
	// usual rules for check status aggregation, so for example if any
	// one step is failing then the entire scenario has failed.
	Status checks.Status

	// Steps describes the results of each of the scenario's test steps.
	Steps []StepResult
}

// StepResult represents the result of executing a single step within a test
// scenario.
type StepResult struct {
	// Name is the user-selected name for the step, or it's a system-generated
	// implied step name which is then guaranteed to start with "<" and end
	// with ">" to allow distinguishing explicit vs. implied steps.
	Name string

	// Status is the aggregate status across all of the checks in this step,
	//
	// If field Diagnostics includes at least one error diagnostic then Status
	// is always checks.StatusError, regardless of the individual check results.
	//
	// This field also takes into account field [ExpectedFailures]: a failure
	// that was expected is counted as if it were passing, and any passing
	// object is treated as a failure, thereby essentially inverting the
	// result of those checks when considered in aggregate.
	//
	// Status unknown represents that the step didn't run to completion but that
	// any partial execution didn't encounter any failures or errors. For
	// example, a step has an unknown result if an earlier step in the same
	// scenario failed and therefore blocked running the remaining steps.
	Status checks.Status

	// Checks describes the results of each of the checkable objects declared
	// in the configuration for this step.
	//
	// Some implied steps don't actually perform normal Terraform plan/apply
	// operations and so do not produce check results. In that case Checks
	// is nil and Status and Diagnostics together describe the outcome of
	// the step.
	//
	// The special implied steps, like the final "terraform destroy" to clean
	// up anything left dangling, are essentially implementation details
	// rather than a real part of the author's test suite, and so UI code may
	// wish to use more muted presentation when reporting them, or perhaps not
	// mention them at all unless they return errors.
	Checks *states.CheckResults

	// Postconditions describes the results of any additional postconditions
	// declared as part of the test step itself, represented as if the
	// test step as a whole were a checkable object with conditions.
	//
	// Postconditions is nil if the step didn't actually include any
	// postconditions, so that e.g. the UI can skip mentioning them in that
	// case.
	//
	// These are separate from [Checks] because they belong only to the
	// test scenario and so would not affect the runtime behavior of the
	// module under test. The results in [Checks] describe results that would
	// also be relevant when using the same module in a non-testing context.
	Postconditions *states.CheckResultObject

	// ExpectedFailures augments field Checks with extra information about the
	// objects that the test author listed in "expected_failures".
	//
	// All members of this map should be [checks.StatusFail] for the overall
	// test step to be considered as passing. If any entries in this map are
	// [checks.StatusPass] then the overall step status is [checks.StatusFail].
	ExpectedFailures addrs.Map[addrs.Checkable, checks.Status]

	// Diagnostics reports any diagnostics generated during this step.
	//
	// Diagnostics cannot be unambigously associated with specific checks, so
	// in some cases these diagnostics might be the direct cause of some checks
	// having status error, while in other cases the diagnostics may be totally
	// unrelated to any of the checks and instead describe a more general
	// problem.
	Diagnostics tfdiags.Diagnostics
}

func (sr *StepResult) IsImplied() bool {
	if len(sr.Name) == 0 {
		// Should not be possible because this would not be a valid identifier,
		// but we'll treat it as implied anyway to be robust about it.
		return true
	}
	return sr.Name[0] == '<'
}