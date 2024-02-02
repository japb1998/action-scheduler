package action

import (
	"context"
	"fmt"
	"os"

	"github.com/japb1998/action-scheduler/internal/types"
)

type ActionService struct{}

func New() *ActionService {
	return &ActionService{}
}

// actions. this will also have a collection at some point.
var (
	single_email = types.Action{
		Id:   "1",
		Name: "single_email",
		Arn:  os.Getenv("SINGLE_EMAIL_FUNCTION"),
		Role: os.Getenv("SINGLE_EMAIL_ROLE"),
	}
	mass_email = types.Action{
		Id:   "2",
		Name: "mass_email",
		Arn:  os.Getenv("MASS_EMAIL_ARN"),
		Role: os.Getenv("MASS_EMAIL_ROLE"),
	}
)

var actions = []types.Action{
	single_email,
	mass_email,
}

func (as *ActionService) GetActionByID(ctx context.Context, id string) (types.Action, error) {
	for _, ac := range actions {
		if ac.Id == id {
			return ac, nil
		}
	}
	return types.Action{}, fmt.Errorf("Invalid action ID provided")
}
