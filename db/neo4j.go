package db

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func NewDriver(uri, user, password string) (neo4j.DriverWithContext, error) {
	return neo4j.NewDriverWithContext(uri,
		neo4j.BasicAuth(user, password, ""))
}
