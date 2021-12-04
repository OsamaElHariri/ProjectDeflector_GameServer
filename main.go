package main

import (
	"log"
	gamemechanics "projectdeflector/game/game_mechanics"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	gameStorage := gamemechanics.NewStorage()

	app.Get("/game/:gameId", func(c *fiber.Ctx) error {
		gameId, err := strconv.Atoi(c.Params("gameId"))
		if err != nil {
			return err
		}

		defenition, ok := gameStorage.Get(gameId)
		if !ok {
			return c.SendStatus(400)
		}

		gameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		return c.JSON(parseGameBoard(gameBoard))
	})

	app.Post("/game", func(c *fiber.Ctx) error {
		payload := struct {
			GameId int `json:"gameId"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}

		defenition := gamemechanics.NewGameBoardDefinition()

		gameStorage.Set(payload.GameId, defenition)

		gameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		redVariants := gamemechanics.GetPawnVariants(payload.GameId, gamemechanics.RED_SIDE, 2)
		blueVariants := gamemechanics.GetPawnVariants(payload.GameId, gamemechanics.BLUE_SIDE, 2)

		return c.JSON(fiber.Map{
			"gameBoard":    parseGameBoard(gameBoard),
			"playerTurn":   parsePlayerTurn(gamemechanics.GetPlayerTurn(gameBoard.Turn)),
			"redVariants":  redVariants,
			"blueVariants": blueVariants,
		})
	})

	app.Delete("/game", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	app.Post("/turn", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     int    `json:"gameId"`
			X          int    `json:"x"`
			Y          int    `json:"y"`
			PlayerSide string `json:"playerSide"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		gameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		turnsPlayed := gameBoard.GetTurnsPlayed(payload.PlayerSide)

		var playerId int
		if payload.PlayerSide == "red" {
			playerId = gamemechanics.RED_SIDE
		} else {
			playerId = gamemechanics.BLUE_SIDE
		}

		variants := gamemechanics.GetPawnVariants(payload.GameId, playerId, turnsPlayed+1)
		event := gamemechanics.NewGameEvent(gamemechanics.CREATE_PAWN, payload.X, payload.Y, variants[len(variants)-1])
		gameBoard, _ = gamemechanics.AddEvent(gameBoard, event)

		fireEvent := gamemechanics.NewGameEvent(gamemechanics.FIRE_DEFLECTOR, 0, 0, payload.PlayerSide)
		gameBoard, deflections := gamemechanics.AddEvent(gameBoard, fireEvent)

		gameStorage.Set(payload.GameId, gameBoard.GetDefenition())

		redVariants := gamemechanics.GetPawnVariants(payload.GameId, gamemechanics.RED_SIDE, gameBoard.GetTurnsPlayed("red")+2)
		blueVariants := gamemechanics.GetPawnVariants(payload.GameId, gamemechanics.BLUE_SIDE, gameBoard.GetTurnsPlayed("blue")+2)

		return c.JSON(fiber.Map{
			"gameBoard":    parseGameBoard(gameBoard),
			"deflections":  parseDeflections(deflections),
			"redVariants":  redVariants,
			"blueVariants": blueVariants,
			"playerTurn":   parsePlayerTurn(gamemechanics.GetPlayerTurn(gameBoard.Turn)),
		})
	})

	app.Post("/peek", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     int    `json:"gameId"`
			X          int    `json:"x"`
			Y          int    `json:"y"`
			PlayerSide string `json:"playerSide"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		gameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		var playerId int
		if payload.PlayerSide == "red" {
			playerId = gamemechanics.RED_SIDE
		} else {
			playerId = gamemechanics.BLUE_SIDE
		}

		variants := gamemechanics.GetPawnVariants(payload.GameId, playerId, gameBoard.Turn/2)
		event := gamemechanics.NewGameEvent(gamemechanics.CREATE_PAWN, payload.X, payload.Y, variants[len(variants)-1])
		gameBoard, _ = gamemechanics.AddEvent(gameBoard, event)

		fireEvent := gamemechanics.NewGameEvent(gamemechanics.FIRE_DEFLECTOR, 0, 0, "")
		gameBoard, deflections := gamemechanics.AddEvent(gameBoard, fireEvent)

		return c.JSON(fiber.Map{
			"gameBoard":   parseGameBoard(gameBoard),
			"deflections": parseDeflections(deflections),
		})
	})

	log.Fatal(app.Listen(":3000"))
}
