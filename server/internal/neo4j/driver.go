package neo4j

import (
	"fmt"
	"log"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	driver     neo4j.DriverWithContext
	driverOnce sync.Once
)

func Init(uri, username, password string) error {
	var initErr error
	driverOnce.Do(func() {
		d, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""), func(c *neo4j.Config) {
			c.MaxConnectionPoolSize = 20
		})
		if err != nil {
			initErr = fmt.Errorf("neo4j driver init: %w", err)
			return
		}
		if err := d.VerifyConnectivity(driverCtx()); err != nil {
			initErr = fmt.Errorf("neo4j connectivity check: %w", err)
			return
		}
		driver = d
		log.Println("[neo4j] connected")
	})
	return initErr
}

func Driver() neo4j.DriverWithContext {
	if driver == nil {
		panic("neo4j not initialized")
	}
	return driver
}

func Close() {
	if driver != nil {
		driver.Close(driverCtx())
	}
}
