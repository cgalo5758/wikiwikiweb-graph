package main

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/spf13/viper"
)

// Create  Neo4j Driver
func graphdb() neo4j.Driver {
	driver, err := neo4j.NewDriver(viper.GetString("neo4j.uri"),
		neo4j.BasicAuth(viper.GetString("neo4j.username"),
			viper.GetString("neo4j.password"), ""))
	if err != nil {
		panic(err)
	}
	return driver
}

// Verify Connectivity
func verify(driver neo4j.Driver) error {
	err := driver.VerifyConnectivity()
	return err
}
