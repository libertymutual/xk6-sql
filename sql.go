package sql

import (
	dbsql "database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/snowflakedb/gosnowflake"

	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/sql", new(SQL))
}

// SQL is the k6 SQL plugin.
type SQL struct{}
type keyValue map[string]interface{}

func contains(array []string, element string) bool {
	for _, item := range array {
		if item == element {
			return true
		}
	}
	return false
}

func (*SQL) Open(database string, connectionString string) (*dbsql.DB, error) {
	supportedDatabases := []string{"mysql", "postgres", "sqlite3", "sqlserver", "snowflake"}
	if !contains(supportedDatabases, database) {
		return nil, fmt.Errorf("database %s is not supported", database)
	}

	db, err := dbsql.Open(database, connectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (*SQL) Query(db *dbsql.DB, query string, args ...interface{}) ([]keyValue, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	result := make([]keyValue, 0)

	for rows.Next() {
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)

		if err != nil {
			return nil, err
		}

		data := make(keyValue, len(cols))
		for i, colName := range cols {
			data[colName] = *valuePtrs[i].(*interface{})
		}
		result = append(result, data)
	}

	rows.Close()
	return result, nil
}
