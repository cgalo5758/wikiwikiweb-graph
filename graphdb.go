package main

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/spf13/viper"
)

func check() {
	driver := graphdb()
	defer driver.Close()
	err := verify(driver)
	if err != nil {
		// print error
		fmt.Println("Error:", err)
		panic(err)
	}
	// print error
	fmt.Println("Neo4j is running")
}

// Create  Neo4j Driver
func graphdb() neo4j.Driver {
	// print neo4j config
	fmt.Println("neo4j config:")
	fmt.Println("neo4j.uri:", viper.GetString("neo4j.uri"))
	fmt.Println("neo4j.user:", viper.GetString("neo4j.user"))
	fmt.Println("neo4j.password:", viper.GetString("neo4j.password"))

	driver, err := neo4j.NewDriver(viper.GetString("neo4j.uri"),
		neo4j.BasicAuth(viper.GetString("neo4j.user"),
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

//export to Neo4j
func export(sourceDir, destDir string) {
	driver := graphdb()
	defer driver.Close()
	// err := exportToNeo4j(driver, sourceDir, destDir)
	// if err != nil {
	// 	panic(err)
	// }
}
