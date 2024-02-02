package schedule

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/japb1998/action-scheduler/internal/service/schedule"
	"github.com/japb1998/action-scheduler/internal/types"
)

type ScheduleService interface {
	GetByID(c context.Context, id string) (*types.Schedule, error)
	GetPaginated(c context.Context, pagination *types.PaginationOps) (*types.PaginatedResult[types.Schedule], error)
	Create(c context.Context, schedule types.CreateScheduleInput) (*types.Schedule, error)
	Delete(c context.Context, id string) error
}

func GetSchedules(ctx *gin.Context) {

	var paginationOps types.PaginationOps

	if err := ctx.ShouldBindQuery(&paginationOps); err != nil {
		var e validator.ValidationErrors

		if errors.As(err, &e) {
			errSlice := make([]struct {
				Field string
				Error string
			}, 0)
			for _, err := range e {
				errSlice = append(errSlice, struct {
					Field string
					Error string
				}{
					Error: err.Error(),
					Field: err.Field(),
				})
			}

			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errors": e,
			})
			return
		}

	}

	if paginationOps.Limit == 0 {
		paginationOps.Limit = 10
	}

	schedules, err := scheduleSvc.GetPaginated(ctx.Request.Context(), &paginationOps)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, schedules)

}

func CreateSchedule(ctx *gin.Context) {
	var sch types.CreateScheduleInput
	if err := ctx.BindJSON(&sch); err != nil {
		var e validator.ValidationErrors
		log.Println(err)
		if errors.As(err, &e) {
			errSlice := make([]struct {
				Field string `json:"field"`
				Error string `json:"error"`
			}, 0)
			for _, err := range e {
				errSlice = append(errSlice, struct {
					Field string `json:"field"`
					Error string `json:"error"`
				}{
					Error: err.Error(),
					Field: err.Field(),
				})
			}

			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errors": errSlice,
			})
			return
		} else {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	newSch, err := scheduleSvc.Create(ctx.Request.Context(), sch)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, newSch)
}

func GetScheduleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	sch, err := scheduleSvc.GetByID(ctx.Request.Context(), id)

	if err != nil {
		if errors.Is(schedule.ErrScheduleNotFound, err) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("schedule with ID: %s not found", id),
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, sch)
}

func DeleteSchedule(ctx *gin.Context) {
	id := ctx.Param("id")

	err := scheduleSvc.Delete(ctx.Request.Context(), id)

	if err != nil {
		if errors.Is(schedule.ErrScheduleNotFound, err) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("schedule with ID: %s not found", id),
			})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
