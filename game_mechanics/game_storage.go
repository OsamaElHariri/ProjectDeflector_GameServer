package gamemechanics

type GameStorage struct {
	games map[int]GameBoardDefenition
}

func NewStorage() GameStorage {
	return GameStorage{
		games: make(map[int]GameBoardDefenition),
	}
}

func (storage *GameStorage) Set(id int, defenition GameBoardDefenition) bool {
	storage.games[id] = defenition
	return true
}

func (storage *GameStorage) Get(id int) (GameBoardDefenition, bool) {
	board, ok := storage.games[id]
	return board, ok
}

func (storage *GameStorage) Delete(id int) bool {
	delete(storage.games, id)
	return true
}
