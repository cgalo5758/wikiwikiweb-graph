package main

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Create  Neo4j Driver
func graphdb() neo4j.Driver {
	driver, err := neo4j.NewDriver("bolt://localhost:7687",
		neo4j.BasicAuth("neo4j", "neo4j", ""))
	if err != nil {
		panic(err)
	}
	return driver
}
