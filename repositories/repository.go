package repositories

type RepositoryFactory interface {
	GetRepository() (Repository, func(), error)
}

type Repository interface {
	InsertGame(defenition InserGameBoardDefenition) (string, error)
	ReplaceGame(objectId string, defenition InserGameBoardDefenition) error
	GetGame(id string) (GetGameBoardDefenitionResult, error)
	GetPlayersGameStats(playerIds []string) ([]PlayerGameStats, error)
}

func GetRepositoryFactory() RepositoryFactory {
	return getMongoRepositoryFactory()
}
