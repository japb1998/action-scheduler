package mapper

import (
	"fmt"
	"time"

	"github.com/japb1998/action-scheduler/internal/model"
	"github.com/japb1998/action-scheduler/internal/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MapScheduleModelToType maps schedule model -> types
func MapScheduleModelToType(model *model.Schedule, action types.Action) *types.Schedule {
	return &types.Schedule{
		ID:          model.ID.Hex(),
		Payload:     model.Payload,
		CreatedBy:   model.CreatedBy,
		Action:      action,
		ClientToken: model.ClientToken,
		Name:        model.Name,
		Expression:  MapModelExpressionToType(model.Expression),
	}
}

// MapModelExpressionToType maps expression model -> types
func MapModelExpressionToType(model model.Expression) types.Expression {
	return types.Expression{
		Start: types.ScheduleDate(model.Start),
		End:   types.ScheduleDate(model.End),
		Type:  string(model.Type),
	}
}

// MapTypeExpressionToModel maps expression types -> model
func MapTypeExpressionToModel(te types.Expression) model.Expression {
	return model.Expression{
		Start: time.Time(te.Start),
		End:   time.Time(te.End),
		Type:  model.ExpressionType(te.Type),
	}
}

// MapScheduleTypesToModelList maps schedule types list -> model list
func MapScheduleTypesToModelList(typesList []types.Schedule) ([]model.Schedule, error) {
	modelList := make([]model.Schedule, 0, len(typesList))

	for _, types := range typesList {
		model, err := MapScheduleTypesToModel(&types)

		if err != nil {
			return nil, fmt.Errorf("failed to map schedule types to model. error=%w", err)
		}
		modelList = append(modelList, *model)
	}

	return modelList, nil
}

// MapScheduleTypesToModel maps schedule types -> model
func MapScheduleTypesToModel(types *types.Schedule) (*model.Schedule, error) {

	id, err := primitive.ObjectIDFromHex(types.ID)

	if err != nil {
		return nil, fmt.Errorf("invalid id on schedule ID field. error=%w", err)
	}

	return &model.Schedule{
		ID:          id,
		Payload:     types.Payload,
		CreatedBy:   types.CreatedBy,
		ActionID:    types.Action.Id,
		ClientToken: types.ClientToken,
		Name:        types.Name,
		Expression:  MapTypeExpressionToModel(types.Expression),
	}, nil
}
