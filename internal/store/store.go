package store

import "database/sql"

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	err = database.Ping()
	if err != nil {
		return nil, err
	}
	return &Store{db: database}, nil
}
 
func migrate(database *sql.DB) error {
	query = "CREATE TABLE IF NOT EXISTS budget_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slo_name    TEXT NOT NULL,
    burn_rate   REAL NOT NULL,
    budget_remaining REAL NOT NULL,
    error_rate  REAL NOT NULL,
    recorded_at DATETIME NOT NULL
);
"
	result, err := database.Exec()
}
