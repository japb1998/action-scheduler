package schedule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"log/slog"

	"github.com/google/uuid"
	"github.com/japb1998/action-scheduler/internal/mapper"
	"github.com/japb1998/action-scheduler/internal/model"
	"github.com/japb1998/action-scheduler/internal/store"
	"github.com/japb1998/action-scheduler/internal/types"
	"github.com/japb1998/action-scheduler/pkg/scheduler"
)

var (
	ErrScheduleNotFound = fmt.Errorf("schedule not found")
	ErrorInvalidPayload = errors.New("invalid payload")
)

type SchedulerStore interface {
	GetByID(context.Context, string) (*model.Schedule, error)
	Get(context.Context, *types.PaginationOps) (int64, []model.Schedule, error)
	Create(ctx context.Context, schedule *model.CreateScheduleInput) (string, error)
	Delete(c context.Context, id string) error
	Update(c context.Context, id string, notification model.Schedule) (*model.Schedule, error)
}

type ActionSvc interface {
	GetActionByID(ctx context.Context, id string) (types.Action, error)
}

type SchedulerService struct {
	// repository
	store     SchedulerStore
	actionSvc ActionSvc
	scheduler scheduler.Scheduler
	logger    *slog.Logger
}

func New(s SchedulerStore, actionSvc ActionSvc, scheduler scheduler.Scheduler) *SchedulerService {
	return &SchedulerService{
		store:     s,
		scheduler: scheduler,
		actionSvc: actionSvc,
		logger:    slog.New(slog.NewTextHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("service", "notification")})),
	}
}

func (s *SchedulerService) GetByID(c context.Context, id string) (*types.Schedule, error) {
	s.logger.Info("getting schedule by ID", "id", id)
	modelS, err := s.store.GetByID(c, id)

	if err != nil {
		s.logger.Error("error getting schedule by ID", "id", id, "error", err.Error())
		if errors.Is(store.ErrScheduleNotFound, err) {
			return nil, ErrScheduleNotFound
		}
		return nil, fmt.Errorf("failed to get schedule with ID='%s'", id)
	}

	action, err := s.actionSvc.GetActionByID(c, modelS.ActionID)

	if err != nil {
		s.logger.Error("error finding action", "action_id", modelS.ActionID, "error", err.Error())
		return nil, err
	}

	return mapper.MapScheduleModelToType(modelS, action), nil
}

func (s *SchedulerService) GetPaginated(c context.Context, pagination *types.PaginationOps) (*types.PaginatedResult[types.Schedule], error) {
	s.logger.Info("getting schedules", "pagination", pagination)

	if pagination == nil {
		return nil, fmt.Errorf("Invalid pagination provider. got=%v", pagination)
	}

	count, models, err := s.store.Get(c, pagination)

	if err != nil {
		s.logger.Error("error getting schedules", slog.String("error", err.Error()))
		return nil, fmt.Errorf("error getting schedules")
	}
	schedules := make([]types.Schedule, 0, len(models))
	for _, schedule := range models {

		s.logger.Info("getting action for schedule", "scheduleID", schedule.ID, "actionID", schedule.ActionID)
		action, err := s.actionSvc.GetActionByID(c, schedule.ActionID)

		if err != nil {
			s.logger.Error("failed to get Action By ID", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to get schedules.")
		}
		schedules = append(schedules, *mapper.MapScheduleModelToType(&schedule, action))
	}
	s.logger.Info("succeeded to get schedules", "count", len(schedules))
	return &types.PaginatedResult[types.Schedule]{
		Total: int(count),
		Items: schedules,
		Limit: pagination.Limit,
		Page:  pagination.Page,
	}, nil
}

func (s *SchedulerService) Create(c context.Context, schedule types.CreateScheduleInput) (sch *types.Schedule, err error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("recover from panic", "recover", r)
			err = fmt.Errorf("unknown failure creating schedule")
		}
	}()
	action, err := s.actionSvc.GetActionByID(c, schedule.ActionID)

	if err != nil {
		return nil, err
	}

	ct := uuid.New()

	cs := model.CreateScheduleInput{
		Expression:  mapper.MapTypeExpressionToModel(schedule.Expression),
		Payload:     schedule.Payload,
		CreatedBy:   schedule.CreatedBy,
		ActionID:    schedule.ActionID,
		Name:        schedule.Name,
		ClientToken: ct.String(),
	}

	schedulerExpression, err := scheduler.NewExpression(cs.Start, cs.End, string(cs.Expression.Type))

	if err != nil {
		return nil, err
	}
	// TODO: input validation before creating schedule
	// the payload validation should be based on the action type
	by, err := json.Marshal(schedule.Payload)

	if err != nil {
		return nil, ErrorInvalidPayload
	}
	// schedules name should be unique
	id, err := s.store.Create(c, &cs)
	scheduleName := fmt.Sprintf("%s-%s", cs.Name, id)
	if err != nil {
		s.logger.Error("error creating schedule", slog.String("error", err.Error()))

		return nil, err
	}

	schedulerInput := scheduler.NewSchedule(scheduleName, action.Arn, action.Role, scheduler.TimeZoneETD, string(by), *schedulerExpression)

	_, err = s.scheduler.CreateSchedule(schedulerInput, cs.ClientToken)

	if err != nil {
		s.logger.Error("error creating eb schedule", slog.String("error", err.Error()))
		err = s.store.Delete(c, id)

		if err != nil {
			s.logger.Error("error deleting schedule from DB", slog.String("error", err.Error()))
			/* TODO: retry. if fails again take action.*/
			return nil, fmt.Errorf("error creating schedule. schedule may have ")
		}
		/*
			TODO: delete schedule from scheduler
		*/
		return nil, err
	}

	if createdModel, err := s.store.GetByID(c, id); err != nil {
		s.logger.Error("error getting schedule", slog.String("error", err.Error()))
		return nil, err
	} else {

		return mapper.MapScheduleModelToType(createdModel, action), nil
	}
}

func (s *SchedulerService) Delete(c context.Context, id string) error {
	s.logger.Info("deleting schedule", "id", id)
	modelS, err := s.store.GetByID(c, id)

	if err != nil {
		s.logger.Error("error getting schedule by ID", "id", id, "error", err.Error())
		if errors.Is(store.ErrScheduleNotFound, err) {
			return ErrScheduleNotFound
		}
		return fmt.Errorf("failed to get schedule with ID='%s'", id)
	}

	err = s.scheduler.DeleteSchedule(fmt.Sprintf("%s-%s", modelS.Name, modelS.ID.Hex()), modelS.ClientToken)

	if err != nil {
		s.logger.Error("error deleting schedule", "id", id, "error", err.Error())
		if errors.Is(scheduler.ErrNotFound, err) {
			s.logger.Error("schedule not found in scheduler", "id", id, "error", err.Error())
			s.logger.Error("deleting schedule from DB", "id", id)
		} else {
			return fmt.Errorf("failed to delete schedule with ID='%s'", id)
		}
	}

	err = s.store.Delete(c, id)

	if err != nil {
		s.logger.Error("error deleting schedule", "id", id, "error", err.Error())
		return fmt.Errorf("failed to delete schedule with ID='%s'", id)
	}

	return nil
}
