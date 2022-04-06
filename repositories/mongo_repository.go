package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	TimePerTurn int      `bson:"time_per_turn"`
	StartTime   int64    `bson:"start_time"`
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
	TimePerTurn int      `bson:"time_per_turn"`
	StartTime   int64    `bson:"start_time"`
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

func (repo MongoRepository) GetOngoingPlayerGame(playerId string) (GetGameBoardDefenitionResult, error) {
	var result GetGameBoardDefenitionResult

	filter := bson.D{
		{Key: "player_ids", Value: playerId},
		{Key: "winner", Value: ""},
	}
	err := repo.client.Database("game_management").Collection("games").FindOne(repo.ctx, filter).Decode(&result)

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
		Wins:     0,
		Games:    0,
	}

	if len(results) > 0 {
		stats.Wins = int(results[0]["wins"].(int32))
		stats.Games = int(results[0]["totalGames"].(int32))
	}

	return stats, nil
}

type WinStreak struct {
	HasWonToday bool
	WinStreak   int
	NextDay     int64
}

func (repo MongoRepository) GetWinStreak(playerId string) (WinStreak, error) {
	gameTimes, err := getWonGamesStartTimes(repo, playerId)
	if err != nil {
		return WinStreak{}, err
	}
	hasWonToday := false
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayCheck := today
	totalDayStreak := 0
	for i := 0; i < len(gameTimes); i++ {
		gameTime := time.UnixMilli(gameTimes[i].StartTime)
		gameTime = time.Date(gameTime.Year(), gameTime.Month(), gameTime.Day(), 0, 0, 0, 0, gameTime.Location())
		if gameTime.Equal(today) {
			hasWonToday = true
			continue
		}

		if gameTime.Before(dayCheck) {
			if dayCheck.Sub(gameTime).Hours() > 24 {
				break
			} else {
				dayCheck = gameTime
				totalDayStreak += 1
			}
		}
	}

	if hasWonToday {
		totalDayStreak += 1
	}
	return WinStreak{
		HasWonToday: hasWonToday,
		WinStreak:   totalDayStreak,
		NextDay:     time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).UnixMilli(),
	}, nil
}

type GameTime struct {
	StartTime int64 `bson:"start_time"`
}

func getWonGamesStartTimes(repo MongoRepository, playerId string) ([]GameTime, error) {
	filter := bson.D{
		{Key: "winner", Value: playerId},
	}
	opt := options.Find()
	opt.Projection = bson.D{
		{Key: "start_time", Value: 1},
	}

	opt.Sort = bson.D{
		{Key: "start_time", Value: -1},
	}

	cursor, err := repo.client.Database("game_management").Collection("games").Find(repo.ctx, filter, opt)

	if err != nil {
		return []GameTime{}, err
	}
	var results []GameTime
	if err = cursor.All(repo.ctx, &results); err != nil {
		return []GameTime{}, nil
	}

	if err := cursor.Close(repo.ctx); err != nil {
		return []GameTime{}, nil
	}

	return results, nil
}
