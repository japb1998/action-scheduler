package schedule

import (
	"log/slog"

	"github.com/japb1998/action-scheduler/internal/database/mongodb"
	"github.com/japb1998/action-scheduler/internal/service/action"
	"github.com/japb1998/action-scheduler/internal/service/schedule"
	"github.com/japb1998/action-scheduler/internal/store"
	"github.com/japb1998/action-scheduler/pkg/awssess"
	"github.com/japb1998/action-scheduler/pkg/scheduler"
)

var scheduleSvc ScheduleService

func init() {
	slog.Info("Initializing Schedule Controllers", "package", "schedule")
	// clients
	c := mongodb.MustInit()
	sess := awssess.MustGetSession()

	// dependencies
	schStorage := store.NewMongoScheduleStore(c)
	scheduler := scheduler.NewScheduler(sess, &scheduler.SchedulerOps{
		RetryAttempts: 0,
	})
	actionSvc := action.New()

	scheduleSvc = schedule.New(schStorage, actionSvc, scheduler)
	slog.Info("Schedule Controllers Initialized", "package", "schedule")
}
