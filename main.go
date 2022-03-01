package main

import (
	"log"
	"math/rand"
	broadcast "projectdeflector/game/broadcast"
	player_colors "projectdeflector/game/colors"
	gamemechanics "projectdeflector/game/game_mechanics"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	gameStorage := gamemechanics.NewStorage()

	colorMap := map[string]string{}

	app.Get("/colors/:id", func(c *fiber.Ctx) error {

		playerId := c.Params("id")
		colors := player_colors.GetPlayerColors(playerId, 4)

		return c.JSON(fiber.Map{
			"colors": colors,
		})
	})

	app.Post("/color", func(c *fiber.Ctx) error {
		payload := struct {
			PlayerId string `json:"playerId"`
			Color    string `json:"color"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}
		colorMap[payload.PlayerId] = payload.Color
		return c.JSON(fiber.Map{
			"color": payload.Color,
		})
	})

	app.Get("/game/:id", func(c *fiber.Ctx) error {

		gameId := c.Params("id")
		defenition, ok := gameStorage.Get(gameId)
		if !ok {
			return c.SendStatus(400)
		}

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		fireEvent := gamemechanics.NewFireDeflectorEvent()
		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{fireEvent})
		if err != nil {
			return err
		}

		colors := map[string]string{}
		for _, id := range processedGameBoard.GameBoard.GetDefenition().PlayerIds {
			if val, ok := colorMap[id]; ok {
				colors[id] = val
			} else {
				colors[id] = player_colors.GetPlayerColors(id, 1)[0]
			}
		}

		result := fiber.Map{
			"gameId":            defenition.Id,
			"playerIds":         defenition.PlayerIds,
			"gameBoard":         parseGameBoard(processedGameBoard.GameBoard),
			"playerTurn":        gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard),
			"variants":          processedGameBoard.PawnVariants,
			"targetScore":       defenition.TargetScore,
			"matchPointPlayers": processedGameBoard.PlayersInMatchPoint,
			"colors":            colors,
			"deflections":       parseDeflections(processedGameBoard.LastDeflections),
		}
		return c.JSON(result)
	})

	app.Post("/game", func(c *fiber.Ctx) error {
		payload := struct {
			PlayerIds []string `json:"playerIds"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(400)
		}

		if len(payload.PlayerIds) != 2 {
			return c.SendStatus(400)
		}
		gameId := strconv.Itoa(rand.Int())

		defenition := gamemechanics.NewGameBoardDefinition(gameId, payload.PlayerIds)
		gameStorage.Set(gameId, defenition)

		result := fiber.Map{
			"gameId": gameId,
		}
		return c.JSON(result)
	})

	app.Delete("/game", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	app.Post("/pawn", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     string `json:"gameId"`
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

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		pawnEvent := gamemechanics.NewCreatePawnEvent(gamemechanics.NewPosition(payload.X, payload.Y), payload.PlayerSide)
		var newEvents []gamemechanics.GameEvent

		newEvents = append(newEvents, pawnEvent)

		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, newEvents)

		if err != nil {
			return c.SendStatus(400)
		}

		newPawn, err := processedGameBoard.GameBoard.GetPawn(gamemechanics.NewPosition(payload.X, payload.Y))
		if err != nil {
			return c.SendStatus(400)
		}

		gameStorage.Set(payload.GameId, processedGameBoard.GameBoard.GetDefenition())

		result := fiber.Map{
			"scoreBoard": processedGameBoard.GameBoard.ScoreBoard,
			"variants":   processedGameBoard.PawnVariants,
			"newPawn":    parsePawn(*newPawn),
		}
		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "pawn", result)

		return c.JSON(result)
	})

	app.Post("/turn", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     string `json:"gameId"`
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
		fireEvent := gamemechanics.NewFireDeflectorEvent()
		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{fireEvent})

		if err != nil {
			return err
		}

		allDeflectionsParsed := make([][]Deflection, 0)
		for i := 0; i < len(allDeflections); i++ {
			allDeflectionsParsed = append(allDeflectionsParsed, parseDeflections(allDeflections[i]))
		}

		result := fiber.Map{
			"scoreBoard":        processedGameBoard.GameBoard.ScoreBoard,
			"variants":          processedGameBoard.PawnVariants,
			"playerTurn":        gamemechanics.GetPlayerTurn(processedGameBoard.GameBoard),
			"allDeflections":    allDeflectionsParsed,
			"winner":            processedGameBoard.Winner,
			"matchPointPlayers": processedGameBoard.PlayersInMatchPoint,
			"deflections":       parseDeflections(processedGameBoard.LastDeflections),
		}
		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "turn", result)

		return c.JSON(result)
	})

	app.Post("/shuffle", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     string `json:"gameId"`
			HasPeek    bool   `json:"hasPeek"`
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

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		skipEvent := gamemechanics.NewSkipPawnEvent(payload.PlayerSide)
		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, []gamemechanics.GameEvent{skipEvent})
		if err != nil {
			return c.SendStatus(400)
		}

		gameStorage.Set(payload.GameId, processedGameBoard.GameBoard.GetDefenition())

		result := fiber.Map{
			"hasPeek": payload.HasPeek,
		}

		tempVariants := map[string][]string{}
		for key, val := range processedGameBoard.PawnVariants {
			if key == payload.PlayerSide {
				tempVariants[key] = append([]string{}, val...)
			} else {
				tempVariants[key] = val
			}
		}

		result["variants"] = tempVariants

		if payload.HasPeek {
			peekPosition := gamemechanics.NewPosition(payload.X, payload.Y)
			pawnEvent := gamemechanics.NewCreatePawnEvent(peekPosition, payload.PlayerSide)
			fireEvent := gamemechanics.NewFireDeflectorEvent()
			var newEvents []gamemechanics.GameEvent
			newEvents = append(newEvents, pawnEvent)
			newEvents = append(newEvents, fireEvent)

			processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, newEvents)
			if err != nil {
				return c.SendStatus(400)
			}

			newPawn, err := processedGameBoard.GameBoard.GetPawn(gamemechanics.NewPosition(payload.X, payload.Y))
			if err != nil {
				return c.SendStatus(400)
			}
			result["newPawn"] = parsePawn(*newPawn)
			result["deflections"] = parseDeflections(processedGameBoard.LastDeflections)
		}

		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "shuffle", result)

		return c.JSON(result)
	})

	app.Post("/peek", func(c *fiber.Ctx) error {
		payload := struct {
			GameId     string `json:"gameId"`
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

		processedGameBoard, err := gamemechanics.NewGameBoard(defenition)

		if err != nil {
			return err
		}

		peekPosition := gamemechanics.NewPosition(payload.X, payload.Y)
		pawnEvent := gamemechanics.NewCreatePawnEvent(peekPosition, payload.PlayerSide)
		fireEvent := gamemechanics.NewFireDeflectorEvent()
		var newEvents []gamemechanics.GameEvent
		newEvents = append(newEvents, pawnEvent)
		newEvents = append(newEvents, fireEvent)

		processedGameBoard, err = gamemechanics.ProcessEvents(processedGameBoard, newEvents)

		if err != nil {
			return c.SendStatus(400)
		}

		newPawn, err := processedGameBoard.GameBoard.GetPawn(gamemechanics.NewPosition(payload.X, payload.Y))
		if err != nil {
			return c.SendStatus(400)
		}

		result := fiber.Map{
			"newPawn":     parsePawn(*newPawn),
			"deflections": parseDeflections(processedGameBoard.LastDeflections),
		}
		broadcast.SocketBroadcast(processedGameBoard.GameBoard.GetDefenition().PlayerIds, "peek", result)

		return c.JSON(result)
	})

	log.Fatal(app.Listen(":3000"))
}
