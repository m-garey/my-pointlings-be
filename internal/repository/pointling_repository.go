package pointling_repo

type PointlingRepository struct {
	db *sql.DB
}

type API interace {

}

func New(db sql.DB) *PointlingRepository {
	return &PointlingRepository{db: db}
}
