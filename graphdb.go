package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/spf13/viper"
)

func check() {
	driver := getDriver()
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
func getDriver() neo4j.Driver {
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

// Export to Neo4j
func export(sourceDir string) {
	// Check if sourceDir directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		fmt.Println("Error:", err)
		panic(err)
	}

	// Get list of markdown files in sourceDir
	files, err := getFiles(sourceDir)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	// Get listSize of list and round up to nearest factor of batchSize
	listSize := len(files)
	batchSize := 256
	batches := listSize / batchSize

	if listSize%batchSize != 0 {
		batches++
	}
	listSize = batches * batchSize

	// create a slice of maps to store nodes.
	var nodes []map[string]interface{}

	// Make a slice of string arrays of size 2 to store relationships
	relationships := make([][2]string, 0)

	// For each markdown file
	for _, file := range files {
		// Extract title and store in a node map
		node := getNode(file)
		// insert node to nodes slice
		nodes = append(nodes, node)
		// Build relationship strings from internal links
		nodeRelationships := getRelationships(file)
		// Add relationships to relationships slice
		relationships = append(relationships, nodeRelationships...)
	}

	// adjust capacity of nodes slice to match listSize
	remainderLen := listSize - len(nodes)
	nodes = append(nodes, make([]map[string]interface{}, remainderLen)...)

	// Create a neo4j driver and defer closing it
	driver := getDriver()
	defer driver.Close()

	// Open a new session and defer closing it
	session, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
	defer session.Close()

	// Create a slice of strings to store cypher queries
	queries := make([]string, 0)
	// Create a slice of string arrays to store cypher parameters
	params := make([]map[string]interface{}, 0)

	// Build a cypher query per batch to create nodes
	for batch := 0; batch < batches; batch++ {
		// Build cypher query that:
		// - unwinds the nodes in the current batch
		// - creates a node for each node in the current batch
		// - returns the nodes created
		query := "UNWIND $nodes AS node CREATE (n:Page {title: node.title})"
		// Get parameters for the current batch, without empty nodes
		parameters := map[string]interface{}{"nodes": nodes[batch*batchSize : (batch+1)*batchSize]}
		// Resize parameters to only include nodes which are not empty. Only on last batch
		if batch == batches-1 {
			parameters["nodes"] = parameters["nodes"].([]map[string]interface{})[:len(parameters["nodes"].([]map[string]interface{}))-remainderLen]
		}
		// Print cypher query
		fmt.Println("cypher query:", query)
		// print cypher parameters
		fmt.Println("cypher parameters:", parameters)
		// Add query to queries slice
		queries = append(queries, query)
		// Add parameters to params slice
		params = append(params, parameters)
	}

	// Execute the cypher queries
	for i, query := range queries {
		// Get parameters for the current query
		queryParams := params[i]
		// Execute the cypher query
		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			result, err := tx.Run(query, queryParams)
			if err != nil {
				return nil, err
			}
			return result.Collect()
		})
		if err != nil {
			fmt.Println("Error:", err)
			panic(err)
		}
	}

	// For each batch of relationships
	for i := 0; i < len(relationships); i += 256 {
		// Build a cypher query to:
		//   - match nodes by title in the relationship string array
		//   - merge relationship between matched nodes
		//   - return the merged nodes
		cypher := "UNWIND $source as source" +
			" UNWIND $target as target" +
			" MATCH (source:Page {title: source})" +
			" MATCH (target:Page {title: target})" +
			" MERGE (source)-[r:LINKED_TO]->(target)" +
			" RETURN source, target"
		// Print relationship batch
		fmt.Println("relationship batch:", relationships[i:i+256])
		// Print cypher query
		fmt.Println("cypher query:", cypher)
		// Execute the cypher query with a write transaction
		//   - Handle any errors
		//   - Verify the number of nodes created
		// _, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// 	result, err := tx.Run(cypher, map[string]interface{}{
		// 		"source": relationships[i : i+256][0], "target": relationships[i : i+256][1],
		// 	})
		// 	if err != nil {
		// 		return "", err
		// 	}
		// 	if result.Next() {
		// 		fmt.Println("Relationships created:", result.Record().GetByIndex(0))
		// 	}
		// 	return nil, err
		// })
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// 	panic(err)
		// }
	}

	// Print success message
	fmt.Println("Neo4j import successful")
	// Print number of nodes created
	fmt.Println("Nodes created:", len(nodes))
	// Print number of relationships created
	fmt.Println("Relationships created:", len(relationships))
	// Print number of files processed
	fmt.Println("Files processed:", len(files))
}

// Get list of markdown files in sourceDir
func getFiles(sourceDir string) ([]string, error) {
	// Get list of markdown files in sourceDir
	files, err := filepath.Glob(sourceDir + "/*.md")
	if err != nil {
		return nil, err
	}
	return files, nil
}

// Get node map
func getNode(file string) map[string]interface{} {
	// Extract title and store in a node map
	node := make(map[string]interface{})
	node["title"] = getTitle(file)
	return node
}

// Get title: filename minus .md
func getTitle(file string) string {
	// Remove directory path from filename
	filename := filepath.Base(file)
	// Extract filename and title and store in a node map
	title := filename[:len(filename)-3]
	// Remove directory from title
	return title
}

// Get relationships: internal links
func getRelationships(file string) [][2]string {
	links := getLinks(file)
	relationships := make([][2]string, 0)
	for _, link := range links {
		relationships = append(relationships, [2]string{getTitle(file), getTitle(link)})
	}
	return relationships
}

// Get links: Parse and match all internal links in markdown file
func getLinks(file string) []string {
	// Open file
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// Read file
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	// Close file
	if err := f.Close(); err != nil {
		panic(err)
	}

	// Parse markdown text
	markdown := string(b)
	// Match all links
	re := regexp.MustCompile(`\[.*?\]\(.*?\)`)
	links := re.FindAllString(markdown, -1)
	// Keep only destination of link
	re = regexp.MustCompile(`\(.*?\)`)
	for i, link := range links {
		links[i] = re.FindString(link)
		links[i] = links[i][1 : len(links[i])-1]
	}
	// Remove external links with http uris
	re = regexp.MustCompile(`\[.*?\]\(http.*?\)`)
	links = re.FindAllString(markdown, -1)
	for i, link := range links {
		links[i] = link[1 : len(link)-1]
	}
	return links
}

// Clear database: Delete all nodes and relationships in neo4j database
func clearDatabase() {
	// Get driver
	driver := getDriver()
	// Open session
	session, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
	defer session.Close()

	// Build a cypher query to:
	//   - delete all nodes and relationships
	//   - return the deleted nodes and relationships
	cypher := "MATCH (n) DETACH DELETE n"

	// Execute the cypher query with a write transaction
	//   - Handle any errors
	//   - Verify the number of nodes and relationships deleted
	_, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(cypher, nil)
		if err != nil {
			return "Sorry, it didn't work", err
		}
		if result.Next() {
			fmt.Println("Nodes deleted:", result.Record().GetByIndex(0))
			fmt.Println("Relationships deleted:", result.Record().GetByIndex(1))
		}
		return nil, err
	})
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}

	// Print success message
	fmt.Println("Neo4j clear successful")
}
