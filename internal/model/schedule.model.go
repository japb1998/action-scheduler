package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExpressionType string

const (
	MonthlyExpression ExpressionType = "monthly"
	DailyExpression   ExpressionType = "daily"
	OneTimeExpression ExpressionType = "one_time"
)

type Schedule struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Expression  `json:"expression" bson:"expression"`
	Payload     map[string]any `json:"payload" bson:"payload"`
	CreatedBy   string         `json:"created_by" bson:"created_by"`
	ActionID    string         `json:"action_id" bson:"action"` // action ID
	ClientToken string         `json:"-" bson:"client_token"`
	Name        string         `json:"name" bson:"name"`
}

type CreateScheduleInput struct {
	Expression  `json:"expression" bson:"expression"`
	Payload     map[string]any `json:"payload" bson:"payload"`
	CreatedBy   string         `json:"created_by" bson:"created_by"`
	ActionID    string         `json:"action_id" bson:"action"` // action ID
	ClientToken string         `json:"-" bson:"client_token"`
	Name        string         `json:"name" bson:"name"`
}

type Expression struct {
	Start time.Time      `json:"start" bson:"start"`
	End   time.Time      `json:"end" bson:"end"`
	Type  ExpressionType `json:"type" bson:"type"`
}
