package main

import (
	"log"
	broadcast "projectdeflector/game/broadcast"
	gamemechanics "projectdeflector/game/game_mechanics"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	gameStorage := gamemechanics.NewStorage()

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

		deflectionSource := processedGameBoard.VarianceFactory.GenerateDeflectionSource(processedGameBoard.GameBoard, processedGameBoard.GameBoard.Turn)
		result := fiber.Map{
			"playerIds":         defenition.PlayerIds,
			"gameBoard":         parseGameBoard(processedGameBoard.GameBoard),
			"playerTurn":        gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard),
			"variants":          gamemechanics.GetPawnVariants(processedGameBoard),
			"deflectionSource":  parseDirectedPosition(deflectionSource),
			"targetScore":       defenition.TargetScore,
			"matchPointPlayers": processedGameBoard.PlayersInMatchPoint,
		}
		return c.JSON(result)
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

		deflectionSource := processedGameBoard.VarianceFactory.GenerateDeflectionSource(processedGameBoard.GameBoard, processedGameBoard.GameBoard.Turn)

		result := fiber.Map{
			"gameBoard":        parseGameBoard(processedGameBoard.GameBoard),
			"finalDeflections": parseDeflections(processedGameBoard.LastDeflections),
			"variants":         gamemechanics.GetPawnVariants(processedGameBoard),
			"playerTurn":       gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard),
			"deflectionSource": parseDirectedPosition(deflectionSource),
		}
		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "pawn", result)

		return c.JSON(result)
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

			if len(processedGameBoard.LastDeflections) > 1 {
				lastDirection := processedGameBoard.LastDeflections[len(processedGameBoard.LastDeflections)-1].ToDirection
				playerId, ok := gamemechanics.GetPlayerFromDirection(processedGameBoard.GameBoard.GetDefenition(), lastDirection)

				if ok && processedGameBoard.PlayersInMatchPoint[playerId] {
					winEvent := gamemechanics.NewWinEvent(playerId)
					processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{winEvent})

					if err != nil {
						return c.SendStatus(400)
					}
					break
				}
			}

			allDeflections = append(allDeflections, processedGameBoard.LastDeflections)
			isDense = processedGameBoard.GameBoard.IsDense()
		}

		endTurnEvent := gamemechanics.NewEndTurnEvent(payload.PlayerSide)
		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{endTurnEvent})

		if err != nil {
			return c.SendStatus(400)
		}

		if processedGameBoard.GameInProgress {
			matchPointEvents := gamemechanics.GetMatchPointEvents(processedGameBoard)
			processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, matchPointEvents)
		}

		if err != nil {
			return c.SendStatus(400)
		}

		gameStorage.Set(payload.GameId, processedGameBoard.GameBoard.GetDefenition())

		deflectionSource := processedGameBoard.VarianceFactory.GenerateDeflectionSource(processedGameBoard.GameBoard, processedGameBoard.GameBoard.Turn)

		allDeflectionsParsed := make([][]Deflection, 0)
		for i := 0; i < len(allDeflections); i++ {
			allDeflectionsParsed = append(allDeflectionsParsed, parseDeflections(allDeflections[i]))
		}

		result := fiber.Map{
			"gameBoard":         parseGameBoard(processedGameBoard.GameBoard),
			"finalDeflections":  parseDeflections(processedGameBoard.LastDeflections),
			"variants":          gamemechanics.GetPawnVariants(processedGameBoard),
			"playerTurn":        gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard),
			"deflectionSource":  parseDirectedPosition(deflectionSource),
			"allDeflections":    allDeflectionsParsed,
			"winner":            processedGameBoard.Winner,
			"matchPointPlayers": processedGameBoard.PlayersInMatchPoint,
		}
		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "turn", result)

		return c.JSON(result)
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

		peekPosition := gamemechanics.NewPosition(payload.X, payload.Y)
		pawnEvent := gamemechanics.NewCreatePawnEvent(peekPosition, payload.PlayerSide)
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

		result := fiber.Map{
			"peekPosition":     parsePosition(peekPosition),
			"gameBoard":        parseGameBoard(processedGameBoard.GameBoard),
			"finalDeflections": parseDeflections(processedGameBoard.LastDeflections),
		}
		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "peek", result)

		return c.JSON(result)
	})

	log.Fatal(app.Listen(":3000"))
}
