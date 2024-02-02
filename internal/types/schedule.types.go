package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// schedule types the idea is that this will have its own collection at some point.
const (
	MONTHLY = "monthly"
	DAILY   = "daily"
	ONE     = "one_time"
)

/*
Schedule is a struct that represents a schedule in the database
this schedule will also be stored/triggered by AWS Event Bridge Scheduler.
*/
type Schedule struct {
	ID          string                             `json:"id" binding:"required"`
	CreatedBy   string                             `json:"created_by" binding:"required"`
	Name        string                             `json:"name"`
	Payload     map[string]any                     `json:"payload,omitempty" binding:"omitempty"`
	Action      `json:"action" binding:"required"` // arn to the lambda function to be triggered
	Expression  `json:"expression" binding:"required"`
	ClientToken string `json:"-"`
}

type CreateScheduleInput struct {
	CreatedBy  string `json:"created_by" binding:"required"`
	ActionID   string `json:"action" binding:"required"`
	Name       string `json:"name" binding:"required,min=2"`
	Expression `json:"expression" binding:"required"`
	Payload    map[string]any `json:"payload,omitempty" binding:"omitempty"`
}

type UpdateScheduleInput struct {
	ActionID string         `json:"action"`
	Name     string         `json:"name" binding:"omitempty,min=2"`
	Payload  map[string]any `json:"payload,omitempty"`
}

// Schedule expression. expression is used in order to build the schedule and know when it will be triggered
type Expression struct {
	Type  string       `json:"type" binding:"required,oneof=monthly daily one_time"`
	Start ScheduleDate `json:"start_date" binding:"omitempty"`
	End   ScheduleDate `json:"end_date" binding:"omitempty"`
}

type UpdateExpressionInput struct {
	Type  string       `json:"type" binding:"omitempty,oneof=monthly daily one_time"`
	Start ScheduleDate `json:"start_date" binding:"omitempty"`
	End   ScheduleDate `json:"end_date" binding:"omitempty"`
}

type ScheduleDate time.Time

func (s *ScheduleDate) UnmarshalJSON(b []byte) error {
	var d string

	if err := json.Unmarshal(b, &d); err != nil {
		return err
	}

	date, err := time.Parse(time.RFC3339, d)

	if err != nil {
		return fmt.Errorf("invalid date format. wanted=%s. got=%s. error=%w", time.RFC3339, d, err)
	}

	*s = ScheduleDate(date)

	return nil
}
func (s *ScheduleDate) MarshalJSON() ([]byte, error) {
	t := time.Time(*s)

	return json.Marshal(t.Format(time.RFC3339))
}
