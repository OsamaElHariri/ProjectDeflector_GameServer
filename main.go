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

		return c.JSON(parseGameBoard(gameBoard))
	})

	app.Delete("/game", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	app.Post("/turn", func(c *fiber.Ctx) error {
		payload := struct {
			GameId  int    `json:"gameId"`
			X       int    `json:"x"`
			Y       int    `json:"y"`
			Variant string `json:"variant"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		event := gamemechanics.NewGameEvent(payload.X, payload.Y, payload.Variant)
		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		defenition.Events = append(defenition.Events, event)
		gameStorage.Set(payload.GameId, defenition)

		gameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		deflections := gameBoard.GetDeflections(gamemechanics.Position{X: 0, Y: 0}, 0)

		return c.JSON(fiber.Map{
			"gameBoard":   parseGameBoard(gameBoard),
			"deflections": parseDeflections(deflections),
		})
	})

	app.Post("/peek", func(c *fiber.Ctx) error {
		payload := struct {
			GameId  int    `json:"gameId"`
			X       int    `json:"x"`
			Y       int    `json:"y"`
			Variant string `json:"variant"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		event := gamemechanics.NewGameEvent(payload.X, payload.Y, payload.Variant)
		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		defenition.Events = append(defenition.Events, event)

		gameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		deflections := gameBoard.GetDeflections(gamemechanics.Position{X: 0, Y: 0}, 0)

		return c.JSON(fiber.Map{
			"gameBoard":   parseGameBoard(gameBoard),
			"deflections": parseDeflections(deflections),
		})
	})

	log.Fatal(app.Listen(":3000"))
}
