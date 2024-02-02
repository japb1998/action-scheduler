// scheduler package is in charge of scheduling notifications using aws event bridge
package scheduler

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsScheduler "github.com/aws/aws-sdk-go/service/scheduler"
)

// TimeZones our app will support
var (
	TimeZoneETD = "America/New_York"
	TimeZoneECT = "America/Los_Angeles"
)

var (
	ErrInvalidDate       = errors.New("Invalid Date was provided")
	ErrInvalidTZ         = errors.New("Invalid time zone provided")
	ErrNotFound          = errors.New("Schedule Not Found")
	ErrInvalidExpression = errors.New("Invalid Expression Type")
)

type SchedulerOps struct {
	RetryAttempts int64
}
type schedule struct {
	name       string
	timeZone   string
	payload    string
	role       string
	target     string
	expression scheduleExpression
}

type scheduler struct {
	ebScheduler *awsScheduler.Scheduler
	SchedulerOps
}

func NewScheduler(sess *session.Session, ops *SchedulerOps) *scheduler {
	var schedulerOps SchedulerOps
	s := awsScheduler.New(sess)

	if ops == nil {
		schedulerOps = *ops
	}
	return &scheduler{
		ebScheduler:  s,
		SchedulerOps: schedulerOps,
	}
}

func NewSchedule(name, targetID, role, tz, payload string, expression scheduleExpression) *schedule {
	return &schedule{
		name:       name,
		timeZone:   tz,
		payload:    payload,
		role:       role,
		expression: expression, // date in UTC
		target:     targetID,
	}
}

// CreateSchedule creates a schedule using aws eventBridge and returns the schedule name. important: schedule name must be unique.
// token
func (s *scheduler) CreateSchedule(sch *schedule, token string) (name string, err error) {

	var expression string
	var loc *time.Location
	// 1. validate time zone
	loc, err = time.LoadLocation(sch.timeZone)
	// 2. get expression string based on expression type
	expression, err = sch.expression.Expression(loc)

	if err != nil {
		return "", err
	}

	target := &awsScheduler.Target{
		Arn:     &sch.target,
		RoleArn: &sch.role,
		Input:   &sch.payload,
		RetryPolicy: &awsScheduler.RetryPolicy{
			MaximumRetryAttempts: aws.Int64(s.SchedulerOps.RetryAttempts),
		},
	}

	input := &awsScheduler.CreateScheduleInput{
		Name:                       &sch.name,
		ScheduleExpression:         &expression,
		ActionAfterCompletion:      aws.String("DELETE"),
		Target:                     target,
		ScheduleExpressionTimezone: &sch.timeZone,
		FlexibleTimeWindow: &awsScheduler.FlexibleTimeWindow{
			Mode: aws.String("OFF"),
		},
		ClientToken: &token,
	}

	if sch.expression.Type != OneTime && !sch.expression.Start.IsZero() {
		input.StartDate = &sch.expression.Start
	}

	if sch.expression.Type != OneTime && !sch.expression.End.IsZero() {
		input.EndDate = &sch.expression.End
	}

	_, err = s.ebScheduler.CreateSchedule(input)

	if err != nil {
		return "", fmt.Errorf("error while creating schedule error: %w", err)
	}
	return sch.name, nil
}

func (s *scheduler) DeleteSchedule(name, token string) error {
	input := &awsScheduler.DeleteScheduleInput{
		Name:        &name,
		ClientToken: &token,
	}
	_, err := s.ebScheduler.DeleteSchedule(input)
	var notFound *awsScheduler.ResourceNotFoundException
	if errors.As(err, &notFound) {
		return ErrNotFound
	}
	return err
}

func (s *scheduler) GetSchedule(name string) (*schedule, error) {
	input := &awsScheduler.GetScheduleInput{
		Name: aws.String(name),
	}

	output, err := s.ebScheduler.GetSchedule(input)

	if err != nil {
		var notFound *awsScheduler.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	expression, err := unmarshalExpression(*output)

	if err != nil {
		return nil, fmt.Errorf("error while parsing output expression error: %w", err)
	}

	sch := NewSchedule(*output.Name, *output.Target.Arn, *output.Target.RoleArn, *output.ScheduleExpressionTimezone, *output.Target.Input, *expression)

	return sch, nil
}

// loadTz - load time zone or return error if an invalid string is passed.
func loadTz(tz string) (*time.Location, error) {

	switch tz {
	case TimeZoneETD:
		fallthrough
	case TimeZoneECT:
		loc, err := time.LoadLocation(tz)
		return loc, err
	}

	return nil, ErrInvalidTZ
}

// UpdateSchedule - to be implemented
func (s *scheduler) UpdateSchedule(sch *schedule) (name string, err error) {
	return "", nil
}
