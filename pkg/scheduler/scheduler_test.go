package scheduler

import (
	"testing"
	"time"
)

func TestExpression(t *testing.T) {

	expressions := []struct {
		expression string
		scheduleExpression
		valid bool
	}{{
		expression: "cron(0 12 1 * ? *)",
		scheduleExpression: scheduleExpression{
			Start: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			End:   time.Time{},
			Type:  Monthly,
		},
		valid: true,
	}, {
		expression: "rate(1day)",
		scheduleExpression: scheduleExpression{
			Start: time.Time{},
			End:   time.Time{},
			Type:  Daily,
		},
		valid: true,
	},
		{
			expression: "rate(1day)",
			scheduleExpression: scheduleExpression{
				Start: time.Time{},
				End:   time.Now(),
				Type:  Daily,
			},
			valid: false,
		},
	}

	for _, e := range expressions {
		exp, err := NewExpression(e.Start, e.End, e.Type)
		if e.valid && err != nil {
			t.Errorf("error creating expression: %v. got start=%v. end=%v", err, exp.Start, exp.End)
		} else {
			expS, err := exp.Expression(time.UTC)

			if err != nil {
				t.Errorf("error creating expression: %v", err)
			}

			if expS != e.expression {
				t.Errorf("expected expression %s, got %s", e.expression, expS)
			}
		}

		if !e.valid && err == nil {
			t.Errorf("expected error creating expression: %v", err)
		}

	}
}
