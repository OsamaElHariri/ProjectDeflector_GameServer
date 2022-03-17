package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	client *mongo.Client
	ctx    context.Context
}

type InserGameBoardDefenition struct {
	PlayerIds   []string `bson:"player_ids"`
	YMax        int      `bson:"y_max"`
	XMax        int      `bson:"x_max"`
	TargetScore int      `bson:"target_score"`
	Events      []map[string]interface{}
}

func (repo MongoRepository) InsertGame(defenition InserGameBoardDefenition) (string, error) {
	result, err := repo.client.Database("game_management").Collection("games").InsertOne(repo.ctx, defenition)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

type GetGameBoardDefenitionResult struct {
	Id          string   `bson:"_id"`
	PlayerIds   []string `bson:"player_ids"`
	YMax        int      `bson:"y_max"`
	XMax        int      `bson:"x_max"`
	TargetScore int      `bson:"target_score"`
	Events      []map[string]interface{}
}

func (repo MongoRepository) GetGame(id string) (GetGameBoardDefenitionResult, error) {
	var result GetGameBoardDefenitionResult

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return GetGameBoardDefenitionResult{}, err
	}

	filter := bson.D{{Key: "_id", Value: objectId}}
	err = repo.client.Database("game_management").Collection("games").FindOne(repo.ctx, filter).Decode(&result)

	if err != nil {
		return GetGameBoardDefenitionResult{}, err
	}

	return result, nil
}

func (repo MongoRepository) ReplaceGame(id string, defenition InserGameBoardDefenition) error {

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "_id", Value: objectId}}
	_, err = repo.client.Database("game_management").Collection("games").ReplaceOne(repo.ctx, filter, defenition)

	return err
}
