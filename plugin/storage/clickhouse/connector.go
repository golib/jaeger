package clickhouse

import "database/sql"

var (
	defaultConnector Connector = func(cfg *NamespaceConfig) (*sql.DB, error) {
		db, err := sql.Open("clickhouse", cfg.Datasource)
		if err != nil {
			return nil, err
		}

		if err := db.Ping(); err != nil {
			return nil, err
		}

		return db, nil
	}
)

// Connector defines how to connect to the database
type Connector func(cfg *NamespaceConfig) (*sql.DB, error)

func WithConnector(c Connector) {
	defaultConnector = c
}
