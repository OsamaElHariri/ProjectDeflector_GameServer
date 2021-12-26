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

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		return c.JSON(parseGameBoard(processedGameBoard.GameBoard))
	})

	app.Post("/game", func(c *fiber.Ctx) error {
		payload := struct {
			GameId int `json:"gameId"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}

		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			defenition = gamemechanics.NewGameBoardDefinition(payload.GameId)
			gameStorage.Set(payload.GameId, defenition)
		}

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		redVariants := processedGameBoard.VarianceFactory.GeneratePawnVariant(processedGameBoard.GameBoard.GetPlayerDigest("red"), processedGameBoard.GameBoard.GetTurnsPlayed("red")+2)
		blueVariants := processedGameBoard.VarianceFactory.GeneratePawnVariant(processedGameBoard.GameBoard.GetPlayerDigest("blue"), processedGameBoard.GameBoard.GetTurnsPlayed("blue")+2)

		deflectionSource := processedGameBoard.VarianceFactory.GenerateDeflectionSource(processedGameBoard.GameBoard, processedGameBoard.GameBoard.Turn)

		return c.JSON(fiber.Map{
			"gameBoard":        parseGameBoard(processedGameBoard.GameBoard),
			"playerTurn":       parsePlayerTurn(gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard.Turn)),
			"redVariants":      redVariants,
			"blueVariants":     blueVariants,
			"deflectionSource": parseDirectedPosition(deflectionSource),
		})
	})

	app.Delete("/game", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	app.Post("/pawn", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     int    `json:"gameId"`
			X          int    `json:"x"`
			Y          int    `json:"y"`
			PlayerSide string `json:"playerSide"`
			SkipPawn   bool   `json:"skipPawn"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		pawnEvent := gamemechanics.NewCreatePawnEvent(gamemechanics.NewPosition(payload.X, payload.Y), payload.PlayerSide)
		var newEvents []gamemechanics.GameEvent

		if payload.SkipPawn {
			newEvents = append(newEvents, gamemechanics.NewSkipPawnEvent(payload.PlayerSide))
		}

		newEvents = append(newEvents, pawnEvent)

		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, newEvents)

		if err != nil {
			return c.SendStatus(400)
		}

		gameStorage.Set(payload.GameId, processedGameBoard.GameBoard.GetDefenition())

		redVariants := processedGameBoard.VarianceFactory.GeneratePawnVariant(processedGameBoard.GameBoard.GetPlayerDigest("red"), processedGameBoard.GameBoard.GetTurnsPlayed("red")+2)
		blueVariants := processedGameBoard.VarianceFactory.GeneratePawnVariant(processedGameBoard.GameBoard.GetPlayerDigest("blue"), processedGameBoard.GameBoard.GetTurnsPlayed("blue")+2)
		deflectionSource := processedGameBoard.VarianceFactory.GenerateDeflectionSource(processedGameBoard.GameBoard, processedGameBoard.GameBoard.Turn)

		return c.JSON(fiber.Map{
			"gameBoard":        parseGameBoard(processedGameBoard.GameBoard),
			"finalDeflections": parseDeflections(processedGameBoard.LastDeflections),
			"redVariants":      redVariants,
			"blueVariants":     blueVariants,
			"playerTurn":       parsePlayerTurn(gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard.Turn)),
			"deflectionSource": parseDirectedPosition(deflectionSource),
		})
	})

	app.Post("/turn", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     int    `json:"gameId"`
			PlayerSide string `json:"playerSide"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		allDeflections := make([][]gamemechanics.Deflection, 0)

		hasFired := false
		fullOnTurnStart := processedGameBoard.GameBoard.IsFull()

		isDense := true
		for !hasFired || (fullOnTurnStart && isDense) {
			hasFired = true
			fireEvent := gamemechanics.NewFireDeflectorEvent()
			processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{fireEvent})
			if err != nil {
				return c.SendStatus(400)
			}
			allDeflections = append(allDeflections, processedGameBoard.LastDeflections)
			isDense = processedGameBoard.GameBoard.IsDense()
		}

		endTurnEvent := gamemechanics.NewEndTurnEvent(payload.PlayerSide)
		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{endTurnEvent})

		if err != nil {
			return c.SendStatus(400)
		}

		gameStorage.Set(payload.GameId, processedGameBoard.GameBoard.GetDefenition())

		redVariants := processedGameBoard.VarianceFactory.GeneratePawnVariant(processedGameBoard.GameBoard.GetPlayerDigest("red"), processedGameBoard.GameBoard.GetTurnsPlayed("red")+2)
		blueVariants := processedGameBoard.VarianceFactory.GeneratePawnVariant(processedGameBoard.GameBoard.GetPlayerDigest("blue"), processedGameBoard.GameBoard.GetTurnsPlayed("blue")+2)
		deflectionSource := processedGameBoard.VarianceFactory.GenerateDeflectionSource(processedGameBoard.GameBoard, processedGameBoard.GameBoard.Turn)

		allDeflectionsParsed := make([][]Deflection, 0)
		for i := 0; i < len(allDeflections); i++ {
			allDeflectionsParsed = append(allDeflectionsParsed, parseDeflections(allDeflections[i]))
		}

		return c.JSON(fiber.Map{
			"gameBoard":        parseGameBoard(processedGameBoard.GameBoard),
			"finalDeflections": parseDeflections(processedGameBoard.LastDeflections),
			"redVariants":      redVariants,
			"blueVariants":     blueVariants,
			"playerTurn":       parsePlayerTurn(gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard.Turn)),
			"deflectionSource": parseDirectedPosition(deflectionSource),
			"allDeflections":   allDeflectionsParsed,
		})
	})

	app.Post("/peek", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     int    `json:"gameId"`
			X          int    `json:"x"`
			Y          int    `json:"y"`
			PlayerSide string `json:"playerSide"`
			SkipPawn   bool   `json:"skipPawn"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}
		defenition, ok := gameStorage.Get(payload.GameId)
		if !ok {
			return c.SendStatus(400)
		}

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		pawnEvent := gamemechanics.NewCreatePawnEvent(gamemechanics.NewPosition(payload.X, payload.Y), payload.PlayerSide)
		fireEvent := gamemechanics.NewFireDeflectorEvent()
		var newEvents []gamemechanics.GameEvent

		if payload.SkipPawn {
			newEvents = append(newEvents, gamemechanics.NewSkipPawnEvent(payload.PlayerSide))
		}

		newEvents = append(newEvents, pawnEvent)
		newEvents = append(newEvents, fireEvent)

		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, newEvents)

		if err != nil {
			return c.SendStatus(400)
		}

		return c.JSON(fiber.Map{
			"gameBoard":        parseGameBoard(processedGameBoard.GameBoard),
			"finalDeflections": parseDeflections(processedGameBoard.LastDeflections),
		})
	})

	log.Fatal(app.Listen(":3000"))
}
