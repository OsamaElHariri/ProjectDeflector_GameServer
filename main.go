package main

import (
	"log"
	"os"
	gamemechanics "projectdeflector/game/game_mechanics"

	"projectdeflector/game/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	err := godotenv.Load("env/." + env + ".env")
	if err != nil {
		log.Fatalf("could not load env vars ")
	}

	app := fiber.New()
	app.Use(recover.New())

	repoFactory := repositories.GetRepositoryFactory()

	app.Use("/", func(c *fiber.Ctx) error {
		repo, cleanup, err := repoFactory.GetRepository()
		if err != nil {
			return err
		}

		defer cleanup()
		c.Locals("repo", repo)

		return c.Next()
	})

	app.Use("/", func(c *fiber.Ctx) error {
		userId := c.Get("x-user-id")
		if userId != "" {
			c.Locals("userId", userId)
		}
		return c.Next()
	})

	app.Get("/ongoing/game", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		gameId, err := useCase.GetOngoingGameId(playerId)

		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"gameId": gameId,
		})
	})

	app.Get("/game/:id", func(c *fiber.Ctx) error {
		gameId := c.Params("id")

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		processedGameBoard, err := useCase.GetGame(gameId)

		if err != nil {
			return err
		}

		return c.JSON(processedGameBoard.ToMap())
	})

	app.Post("/stats/game", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		stats, err := useCase.GetPlayerStats(playerId)
		if err != nil {
			return err
		}
		return c.JSON(stats.ToMap())
	})

	app.Post("/internal/game", func(c *fiber.Ctx) error {
		payload := struct {
			PlayerIds []string `json:"playerIds"`
		}{}

		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		gameId, err := useCase.CreateNewGame(payload.PlayerIds)
		if err != nil {
			return err
		}

		result := fiber.Map{
			"gameId": gameId,
		}
		return c.JSON(result)
	})

	app.Post("/pawn", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)
		payload := struct {
			GameId string `json:"gameId"`
			X      int    `json:"x"`
			Y      int    `json:"y"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		result, err := useCase.AddPawn(payload.GameId, gamemechanics.AddPawnRequest{
			X:          payload.X,
			Y:          payload.Y,
			PlayerSide: playerId,
		})

		if err != nil {
			return err
		}

		return c.JSON(result.ToMap())
	})

	app.Post("/turn", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)
		payload := struct {
			GameId string `json:"gameId"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		result, err := useCase.EndTurn(payload.GameId, playerId)

		if err != nil {
			return err
		}

		return c.JSON(result.ToMap())
	})

	app.Post("/turn/expire", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)
		payload := struct {
			GameId     string `json:"gameId"`
			EventCount int    `json:"eventCount"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		result, err := useCase.ExpireTurn(payload.GameId, playerId, payload.EventCount)

		if err != nil {
			return err
		}

		return c.JSON(result.ToMap())
	})

	app.Post("/shuffle", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)
		payload := struct {
			GameId string `json:"gameId"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		result, err := useCase.Shuffle(payload.GameId, playerId)

		if err != nil {
			return err
		}

		return c.JSON(result.ToMap())
	})

	app.Post("/peek", func(c *fiber.Ctx) error {
		playerId := c.Locals("userId").(string)
		payload := struct {
			GameId string `json:"gameId"`
			X      int    `json:"x"`
			Y      int    `json:"y"`
		}{}
		if err := c.BodyParser(&payload); err != nil {
			return err
		}

		repo := c.Locals("repo").(repositories.Repository)
		useCase := gamemechanics.UseCase{
			Repo: repo,
		}

		result, err := useCase.Peek(payload.GameId, gamemechanics.PeekRequest{
			X:          payload.X,
			Y:          payload.Y,
			PlayerSide: playerId,
		})

		if err != nil {
			return err
		}

		return c.JSON(result.ToMap())
	})

	log.Fatal(app.Listen(":3000"))
}
