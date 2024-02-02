package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	awsScheduler "github.com/aws/aws-sdk-go/service/scheduler"
)

// schedule types
const (
	Monthly = "monthly"
	Daily   = "daily"
	OneTime = "one_time"
)

var (
	ErrInvalidHour = fmt.Errorf("invalid hour")
	ErrInvalidDay  = fmt.Errorf("invalid day")
)

/*
scheduleExpression
start -

	this is the start time of the schedule. if the schedule type is monthly, this is the day and time of the month.
	if the schedule type is daily, this is the time of day.
	if the schedule type is one_time, this is the date and time of the schedule

end - this is the end time of the schedule. if the schedule type is monthly, this is the end date of the schedule

	if the schedule type is daily, this is the end date of the schedule
	if the schedule type is one_time, this is disregarded

type - this is the type of schedule. it can be monthly, daily, or one_time
*/
type scheduleExpression struct {
	Start time.Time
	End   time.Time
	Type  string
}

// NewExpression creates a new schedule expression, and validates the expression type
func NewExpression(start, end time.Time, t string) (*scheduleExpression, error) {

	exp := &scheduleExpression{
		Start: start,
		End:   end,
		Type:  t,
	}

	return exp, exp.validate()
}

func (se *scheduleExpression) Expression(loc *time.Location) (string, error) {
	switch se.Type {
	case Monthly:
		if se.Start.IsZero() {
			if time.Now().Add(24*time.Hour).In(loc).Month() > time.Now().In(loc).Month() {
				return fmt.Sprintf("cron(%d %d %v * ? *)", time.Now().In(loc).Minute(), time.Now().In(loc).Hour(), "L"), nil
			}
			return fmt.Sprintf("cron(%d %d %d * ? *)", time.Now().In(loc).Minute(), time.Now().In(loc).Hour(), time.Now().In(loc).Day()), nil
		}
		return fmt.Sprintf("cron(%d %d %d * ? *)", se.Start.In(loc).Minute(), se.Start.In(loc).Hour(), se.Start.In(loc).Day()), nil
	case Daily:
		return "rate(1day)", nil
	case OneTime:
		return fmt.Sprintf("at(%s)", se.Start.In(loc).Format("2006-01-02T15:04:05")), nil
	default:
		return "", ErrInvalidExpression
	}
}

func (se *scheduleExpression) String() string {
	return fmt.Sprintf("start: %s, end: %s, type: %s", se.Start, se.End, se.Type)
}

func (se *scheduleExpression) validate() error {
	switch se.Type {
	case OneTime:
		{
			if se.Start.IsZero() {
				return fmt.Errorf("%w. description: start time is required for one time schedule", ErrInvalidExpression)
			}

			if se.Start.Before(time.Now()) {
				return fmt.Errorf("%w. description: start time is before current time", ErrInvalidExpression)
			}
			if !se.End.IsZero() {
				return fmt.Errorf("%w. description: end time is not required for one time schedule", ErrInvalidExpression)
			}
		}
	case Monthly:
		{

			if !se.Start.IsZero() && !se.End.IsZero() && !se.Start.Before(se.End) {
				return fmt.Errorf("%w. description: start time must happen before end", ErrInvalidExpression)
			}
			if se.Start.IsZero() && !se.End.IsZero() && se.End.Before(time.Now().AddDate(0, 1, 0)) {
				return fmt.Errorf("%w. description: date and type combination for this event would never happen", ErrInvalidExpression)
			}
		}
	case Daily:
		{
			if !se.Start.IsZero() && !se.End.IsZero() && !se.Start.Before(se.End) {
				return fmt.Errorf("%w. description: start time must happen before end", ErrInvalidExpression)
			}
			if se.Start.IsZero() && !se.End.IsZero() && se.End.Before(time.Now().Add(24*time.Hour)) {
				return fmt.Errorf("%w. description: end must happen after time:%v", ErrInvalidExpression, se.End)
			}
		}
	default:
		return fmt.Errorf("%w. description: invalid schedule type", ErrInvalidExpression)
	}
	return nil
}

// unmarshalExpression takes a GetScheduleOutput and returns a scheduleExpression or error
func unmarshalExpression(out awsScheduler.GetScheduleOutput) (*scheduleExpression, error) {
	se := &scheduleExpression{}

	expression := *out.ScheduleExpression
	if strings.HasPrefix(expression, "cron") {
		cronFields := strings.Split(strings.TrimSuffix(strings.TrimPrefix(*out.ScheduleExpression, "cron("), ")"), " ")
		hour, err := strconv.Atoi(cronFields[1])

		if err != nil {
			return nil, ErrInvalidHour
		}
		day, err := strconv.Atoi(cronFields[2])
		if err != nil {
			return nil, ErrInvalidDay
		}
		t := time.Date(out.StartDate.Day(), out.StartDate.Month(), day, hour, 0, 0, 0, time.UTC)

		se.Type = Monthly
		se.Start = t
		se.End = *out.EndDate
	} else if strings.HasPrefix(expression, "rate") {
		se.Type = Daily
		se.Start = *out.StartDate
		se.End = time.Time{}
	} else if strings.HasPrefix(expression, "at") {
		se.Type = OneTime
		t, err := time.Parse("2006-01-02T15:04:05", strings.TrimSuffix(strings.TrimPrefix(*out.ScheduleExpression, "at("), ")"))

		if err != nil {
			return nil, fmt.Errorf("error while parsing output expression error: %w", err)
		}
		se.Start = t
	} else {
		return nil, ErrInvalidExpression
	}

	return se, nil
}
