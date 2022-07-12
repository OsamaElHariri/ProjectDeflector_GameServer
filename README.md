# ProjectDeflector GameServer

This repo holds the code to manage games and move them forward for the `Hit Bounce` mobile game.


## Mobile App

The mobile app, and a high level introduction to this project can be found in the [JsClientGame](https://github.com/OsamaElHariri/ProjectDeflector_JsClientGame) repo, which is intended to be the entry point to understanding this project.


## Overview of This Project

This is a Go server that uses the [Fiber](https://gofiber.io/) web framework. The routes are just in `main.go`, and the function of the routes is to only validate the inputs, then call a use case (which is what runs the business logic). The use cases can be found in the `use_cases.go` file, and these should incapsulate all the functions that this server can do.

Note that this project has a `.devcontainer` and is meant to be run inside a dev container.


## Outputting a Binary

To output the binary of this Go code, run the VSCode task using `CTRL+SHIFT+B`. This should be done while inside the dev container.


Once you have the binary, you need to build the docker image _outside_ the dev container. I use this command and just overwrite the image everytime. This keeps the [Infra](https://github.com/OsamaElHariri/ProjectDeflector_Infra) repo simpler.

```
docker build -t project_deflector/game_server:1.0 .
```

## Game State

The game state is an array of arrays. This array holds what is called `Pawns`, which are the pieces that the players put on the board. This array of arrays is not stored, but is constructed on each action. What is stored in the database is an array of event. These events describe the sequence of transformations that an initial game state has went through. So if a player wants to add a pawn, a `CreatePawnEvent` is added to the collection. When the player ends the turn, a `FireDeflectorEvent` event is added, followed by an `EndTurnEvent`.

This way of storing the game state makes it easy to add rules and makes implementing things like game replays easy (although I did not implement game replays in this project, but it wouldn't be too hard to do so).