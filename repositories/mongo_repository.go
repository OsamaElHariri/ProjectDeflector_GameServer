package repositories

import (
	"context"
	"time"

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
	LockUntil   int      `bson:"lock_until"`
	Winner      string
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

func (repo MongoRepository) GetGameAndLock(id string) (GetGameBoardDefenitionResult, error) {
	var result GetGameBoardDefenitionResult

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return GetGameBoardDefenitionResult{}, err
	}

	now := time.Now().Unix()

	// if something is locked for more than 5 seconds
	// it means that something went wrong and it will no longer be locked
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "lock_until", Value: now + 5},
		}},
	}

	filter := bson.D{
		{Key: "_id", Value: objectId},
		{Key: "lock_until", Value: bson.D{
			{Key: "$lte", Value: now},
		}},
	}
	err = repo.client.Database("game_management").Collection("games").FindOneAndUpdate(repo.ctx, filter, update).Decode(&result)

	if err != nil {
		return GetGameBoardDefenitionResult{}, err
	}
	return result, nil
}

func (repo MongoRepository) UnlockGame(id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "_id", Value: objectId}}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "lock_until", Value: 0},
		}},
	}

	_, err = repo.client.Database("game_management").Collection("games").UpdateOne(repo.ctx, filter, update)
	return err
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
	defenition.LockUntil = 0

	_, err = repo.client.Database("game_management").Collection("games").ReplaceOne(repo.ctx, filter, defenition)

	return err
}

type PlayerGameStats struct {
	PlayerId string
	Games    int
	Wins     int
}

func (repo MongoRepository) GetPlayersGameStats(playerIds []string) ([]PlayerGameStats, error) {
	stats := []PlayerGameStats{}

	for i := 0; i < len(playerIds); i++ {
		stat, err := getPlayerGameStats(repo, playerIds[i])
		if err != nil {
			return []PlayerGameStats{}, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func getPlayerGameStats(repo MongoRepository, playerId string) (PlayerGameStats, error) {

	match := bson.D{
		{Key: "player_ids", Value: playerId},
		{Key: "winner", Value: bson.D{
			{Key: "$ne", Value: ""},
			{Key: "$exists", Value: 1},
		}},
	}

	group := bson.D{
		{Key: "_id", Value: nil},
		{Key: "totalGames", Value: bson.D{
			{Key: "$sum", Value: 1},
		}},
		{Key: "wins", Value: bson.D{
			{Key: "$sum", Value: bson.D{
				{Key: "$switch", Value: bson.D{
					{Key: "branches", Value: bson.A{
						bson.D{
							{Key: "case", Value: bson.D{
								{Key: "$eq", Value: bson.A{"$winner", playerId}},
							}},
							{Key: "then", Value: 1},
						},
					}},
					{Key: "default", Value: 0},
				}},
			}},
		}},
	}

	cursor, err := repo.client.Database("game_management").Collection("games").Aggregate(repo.ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: group}},
	})

	if err != nil {
		return PlayerGameStats{}, err
	}
	var results []bson.M
	if err = cursor.All(repo.ctx, &results); err != nil {
		return PlayerGameStats{}, nil
	}

	if err := cursor.Close(repo.ctx); err != nil {
		return PlayerGameStats{}, nil
	}

	stats := PlayerGameStats{
		PlayerId: playerId,
		Wins:     int(results[0]["wins"].(int32)),
		Games:    int(results[0]["totalGames"].(int32)),
	}

	return stats, nil
}
