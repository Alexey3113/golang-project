package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"test3/internal/user"
	"test3/pkg/logging"
)

type db struct {
	collection *mongo.Collection
	logger     *logging.Logger
}

func (d *db) Create(ctx context.Context, user user.User) (string, error) {
	d.logger.Debug("create user")

	result, err := d.collection.InsertOne(ctx, user)
	if err != nil {
		d.logger.Fatalf("Can not create user, due to err: %v", err)
		return "", fmt.Errorf("Can not create user, due to err: %v", err)
	}

	d.logger.Debug("try to convert")
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}

	return "", fmt.Errorf("failder to convert oid(%s) to hex due %v", oid, ok)
}

func (d *db) ReadAll(ctx context.Context) (u []user.User, err error) {
	cursor, err := d.collection.Find(ctx, bson.M{})
	if cursor.Err() != nil {
		return u, fmt.Errorf("failed to find all users in database, error: %v", err)
	}

	if err := cursor.All(ctx, &u); err != nil {
		return u, fmt.Errorf("failed to read all user from cursor, due to error: %v", err)
	}

	return u, nil
}

func (d *db) Read(ctx context.Context, id string) (u user.User, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return u, fmt.Errorf("error to convert id from %s, due to err: %v", id, err)
	}

	filter := bson.M{"_id": oid}

	result := d.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		// TODO ErrEntityNotFound
		return u, fmt.Errorf("error to find user with id(%s), error: %v", filter, result.Err())
	}

	if err := result.Decode(&u); err != nil {
		return u, fmt.Errorf("failed to decode user by id %s, due to error: %v", id, err)
	}
	return u, nil
}

func (d *db) Update(ctx context.Context, user user.User) error {
	objectId, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return fmt.Errorf("error to convert id from %s, due to err: %v", user.ID, err)
	}

	filter := bson.M{
		"_id": objectId,
	}

	userBytes, err := bson.Marshal(user)
	if err != nil {
		return fmt.Errorf("error on marshal user data, due to error: %v", err)
	}

	var updatedUserObj bson.M
	err = bson.Unmarshal(userBytes, &updatedUserObj)
	if err != nil {
		return fmt.Errorf("error on unmarshal user data, due to error: %v", err)
	}

	delete(updatedUserObj, "_id")

	update := bson.M{"$set": updatedUserObj}

	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("error on update one user, due to error %v", err)
	}

	if result.MatchedCount == 0 {
		// TODO ErrEntityNotFound
		return fmt.Errorf("not found")
	}

	d.logger.Tracef("matched %d docs and Modified %d docs", result.MatchedCount, result.ModifiedCount)

	return nil
}

func (d *db) Delete(ctx context.Context, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("error to convert id from %s, due to err: %v", id, err)
	}
	filter := bson.M{"_id": objectId}

	deleteResult, err := d.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("error on delete user with id %s, due to err: %v", id, err)
	}

	if deleteResult.DeletedCount == 0 {
		// TODO ErrEntityNotFound
		return fmt.Errorf("not found")
	}

	d.logger.Tracef("deleted %d docs", deleteResult.DeletedCount)

	return nil
}

func NewStorage(database *mongo.Database, collection string, logger *logging.Logger) user.Storage {
	return &db{
		collection: database.Collection(collection),
		logger:     logger,
	}
}
