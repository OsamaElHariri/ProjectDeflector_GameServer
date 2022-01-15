package gamemechanics

type GameStorage struct {
	games map[string]GameBoardDefenition
}

func NewStorage() GameStorage {
	return GameStorage{
		games: make(map[string]GameBoardDefenition),
	}
}

func (storage *GameStorage) Set(id string, defenition GameBoardDefenition) bool {
	storage.games[id] = defenition
	return true
}

func (storage *GameStorage) Get(id string) (GameBoardDefenition, bool) {
	board, ok := storage.games[id]
	return board, ok
}

func (storage *GameStorage) Delete(id string) bool {
	delete(storage.games, id)
	return true
}
