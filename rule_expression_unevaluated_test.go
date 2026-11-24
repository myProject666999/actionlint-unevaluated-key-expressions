package actionlint

import (
	"strings"
	"testing"
)

func TestRuleExpressionUnevaluatedExpressions(t *testing.T) {
	parse := func(t *testing.T, src string) *Workflow {
		t.Helper()
		w, errs := Parse([]byte(src))
		if len(errs) > 0 {
			t.Fatalf("unexpected parse errors for %q: %v", src, errs)
		}
		if w == nil {
			t.Fatalf("no workflow was parsed for %q", src)
		}
		return w
	}

	run := func(t *testing.T, w *Workflow, cfg *Config) []*Error {
		t.Helper()
		proj := &Project{root: t.TempDir()}
		r := NewRuleExpression(nil, NewLocalReusableWorkflowCache(proj, ".", nil))
		if cfg != nil {
			r.SetConfig(cfg)
		}
		v := NewVisitor()
		v.AddPass(r)
		if err := v.Visit(w); err != nil {
			t.Fatal(err)
		}
		return r.Errs()
	}

	countKind := func(errs []*Error, kind string) int {
		n := 0
		for _, e := range errs {
			if e.Kind == kind {
				n++
			}
		}
		return n
	}

	t.Run("context reference is reported once at first expression start", func(t *testing.T) {
		// Two expressions in one scalar. Only one unevaluated diagnostic is reported, at the first
		// ${{ start.
		w := parse(t, `
on: push
jobs:
  test:
    uses: ${{ github.repository }}/${{ github.ref }}
`)
		errs := run(t, w, nil)
		var ue []*Error
		for _, e := range errs {
			if e.Kind == "unevaluated" {
				ue = append(ue, e)
			}
		}
		if len(ue) != 1 {
			t.Fatalf("expected exactly one unevaluated error but got %d: %v", len(ue), errs)
		}
		if !strings.Contains(ue[0].Message, "jobs.<job_id>.uses") {
			t.Errorf("message must contain the workflow key: %q", ue[0].Message)
		}
		// The first ${{ is at column 11 (": uses: " is 10 chars, then "${{").
		if ue[0].Column != 11 {
			t.Errorf("unevaluated error must be at the first ${{ start (column 11) but got column %d", ue[0].Column)
		}
	})

	t.Run("pure literal expression is not reported", func(t *testing.T) {
		w := parse(t, `
on: push
jobs:
  test:
    uses: ${{ './.github/workflows/x.yml' }}
`)
		errs := run(t, w, nil)
		if n := countKind(errs, "unevaluated"); n != 0 {
			t.Errorf("literal expression must not be reported but got %d unevaluated errors: %v", n, errs)
		}
	})

	t.Run("syntax error suppresses unevaluated for that expression only", func(t *testing.T) {
		w := parse(t, `
on: push
jobs:
  test:
    uses: ${{ 1 + }}-${{ github.ref }}
`)
		errs := run(t, w, nil)
		if n := countKind(errs, "expression"); n == 0 {
			t.Errorf("syntax error must still be reported: %v", errs)
		}
		// The second expression is syntax-valid and references a context, so exactly one
		// unevaluated diagnostic is reported for it.
		if n := countKind(errs, "unevaluated"); n != 1 {
			t.Errorf("expected one unevaluated error for the valid expression but got %d: %v", n, errs)
		}
	})

	t.Run("tentative key is silent by default", func(t *testing.T) {
		w := parse(t, `
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
        shell: ${{ env.SHELL }}
`)
		errs := run(t, w, nil)
		if n := countKind(errs, "unevaluated"); n != 0 {
			t.Errorf("tentative key must be silent by default but got: %v", errs)
		}
	})

	t.Run("tentative key is reported when enabled", func(t *testing.T) {
		w := parse(t, `
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo hi
        shell: ${{ env.SHELL }}
`)
		cfg := &Config{}
		cfg.UnevaluatedExpressions.Tentative = true
		errs := run(t, w, cfg)
		if n := countKind(errs, "unevaluated"); n != 1 {
			t.Errorf("expected one unevaluated error when tentative is enabled but got %d: %v", n, errs)
		}
		// The misleading context availability error must be suppressed.
		if n := countKind(errs, "expression"); n != 0 {
			t.Errorf("context availability error must be suppressed at an unevaluated position but got: %v", errs)
		}
	})
}
