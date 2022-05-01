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

	nodes, relationships := getFilesAsGraph(files)

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

	// Build graph in Neo4j using cypher queries
	// Create nodes
	for _, node := range nodes {
		// Create node
		cypher := fmt.Sprintf("CREATE (n:Page {title:'%s'})", node)
		fmt.Println(cypher)
		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			result, err := tx.Run(cypher, nil)
			if err != nil {
				return nil, err
			}
			return result, nil
		})
		if err != nil {
			fmt.Println("Error:", err)
			panic(err)
		}
	}
	// Create relationships
	for _, relationship := range relationships {
		// Create relationship
		cypher := fmt.Sprintf("MATCH (n:Page {title:'%s'}), (m:Page {title:'%s'}) CREATE (n)-[:LINKS_TO]->(m)", relationship[0], relationship[1])
		fmt.Println(cypher)
		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			result, err := tx.Run(cypher, nil)
			if err != nil {
				return nil, err
			}
			return result, nil
		})
		if err != nil {
			fmt.Println("Error:", err)
			panic(err)
		}
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

func getFilesAsGraph(files []string) ([]string, [][2]string) {
	// Create a slice of strings to store the nodes
	nodes := make([]string, 0)
	// Create a slice of [2]strings to store the relationships
	relationships := make([][2]string, 0) // [source, destination]])

	// For each markdown file
	for _, file := range files {
		// Extract title and store in a node slice
		nodes = append(nodes, getTitle(file))
		// Extract relationships and store in a relationship slice
		relationships = append(relationships, getRelationships(file)...)
	}
	return nodes, relationships
}
