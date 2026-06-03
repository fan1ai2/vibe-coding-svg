package neo4j

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var sessionTimeout = 10 * time.Second

func driverCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), sessionTimeout)
	return ctx
}

func NewSession() neo4j.SessionWithContext {
	return Driver().NewSession(driverCtx(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func NewReadSession() neo4j.SessionWithContext {
	return Driver().NewSession(driverCtx(), neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
}
