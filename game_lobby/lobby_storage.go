package gamelobby

type LobbyStorage struct {
	games map[int]Lobby
}

func (storage LobbyStorage) set(id int, gameBoard Lobby) bool {
	storage.games[id] = gameBoard
	return true
}

func (storage LobbyStorage) get(id int) (Lobby, bool) {
	board, ok := storage.games[id]
	return board, ok
}

func (storage LobbyStorage) delete(id int) bool {
	delete(storage.games, id)
	return true
}
