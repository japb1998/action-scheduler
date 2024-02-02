package store

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/japb1998/action-scheduler/internal/model"
	"github.com/japb1998/action-scheduler/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// errors
var (
	ErrScheduleNotFound = errors.New("schedule not found")
	ErrInvalidID        = errors.New("invalid id")
)

type MongoScheduleStore struct {
	coll   *mongo.Collection
	logger *slog.Logger
}

func NewMongoScheduleStore(c *mongo.Client) *MongoScheduleStore {

	return &MongoScheduleStore{
		coll:   c.Database("notification-scheduler").Collection("schedule"),
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil).WithAttrs([]slog.Attr{slog.String("package", "store"), slog.String("collection", "schedule")})),
	}
}

// Get by Id returns the schedule with the given id
func (s *MongoScheduleStore) GetByID(ctx context.Context, id string) (*model.Schedule, error) {
	bsonId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, ErrInvalidID
	}
	filter := bson.D{bson.E{
		Key:   "_id",
		Value: bsonId,
	},
	}

	var schedule model.Schedule

	err = s.coll.FindOne(ctx, filter).Decode(&schedule)

	if err != nil {
		s.logger.Error("error getting schedule by id", slog.String("id", id), slog.String("error", err.Error()))
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}
	return &schedule, nil
}

// GetAll returns all the schedules with pagination. Pagination is Zero based
func (s *MongoScheduleStore) Get(ctx context.Context, pagination *types.PaginationOps) (count int64, schedules []model.Schedule, err error) {
	ops := options.Find().SetSkip(int64(pagination.Limit * pagination.Page)).SetLimit(int64(pagination.Limit))

	cursor, err := s.coll.Find(ctx, bson.D{}, ops)

	if err != nil {
		s.logger.Error("error getting all schedules", slog.String("error", err.Error()))
		return 0, nil, err
	}

	if err = cursor.All(ctx, &schedules); err != nil {
		s.logger.Error("error getting all schedules", slog.String("error", err.Error()))
		return 0, nil, err
	}

	if count, err = s.coll.CountDocuments(ctx, bson.D{}); err != nil {
		s.logger.Error("error getting all schedules", slog.String("error", err.Error()))
		return 0, nil, err
	}

	return count, schedules, nil
}

func (s *MongoScheduleStore) Delete(ctx context.Context, id string) error {
	bsonId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return ErrInvalidID
	}
	filter := bson.D{bson.E{"_id", bsonId}}

	r, err := s.coll.DeleteOne(ctx, filter)

	if err != nil {
		s.logger.Error("error deleting schedule", slog.String("id", id), slog.String("error", err.Error()))
		return err
	}

	if r.DeletedCount == 0 {
		return ErrScheduleNotFound
	}

	return nil
}

func (s *MongoScheduleStore) Create(ctx context.Context, schedule *model.CreateScheduleInput) (string, error) {
	r, err := s.coll.InsertOne(ctx, schedule)
	if err != nil {
		s.logger.Error("error creating schedule", slog.String("error", err.Error()))
		return "", err
	}

	id, ok := r.InsertedID.(primitive.ObjectID)

	if !ok {
		return "", ErrInvalidID
	}

	return id.Hex(), nil
}

func (s *MongoScheduleStore) Update(c context.Context, id string, notification model.Schedule) (*model.Schedule, error) {
	return nil, nil
}
