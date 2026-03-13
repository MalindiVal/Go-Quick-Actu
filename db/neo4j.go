package db

import (
	"context"
	"goquickactu/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

CONST (
	uri      = "bolt://localhost:7687"
	username = "neo4j"
	password = "Acturoot"
)


func NewDriver() (neo4j.DriverWithContext, error) {
	return neo4j.NewDriverWithContext(uri,
		neo4j.BasicAuth(username, password, ""))
}

func ExecuteQuery(ctx context.Context, driver neo4j.DriverWithContext, query string, params map[string]interface{}) (neo4j.ResultWithContext, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	return session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		return tx.Run(ctx, query, params)
	})

}
