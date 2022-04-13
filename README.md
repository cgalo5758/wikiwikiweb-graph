# WikiWikiWeb Graph

Hello, this is a quick code repo to:
- parse wiki.c2.com pages
- turn all pages into markdown
- make the markdown data useful for ~~obsidian~~ neo4j
  - This last part is yet to be implemented. I just need something that will show me a graph of WikiWikiWeb.

## Notes
- Batteries not included. You need to provide your own neo4j db. Basic auth supported.
  - Loading the whole site will definitely exceed Neo4j's AuraDB free quota for number of relationships.
- An Obsidian graph can't handle all WikiWikiWeb pages on most computers. You'd need a very very very beefy computer to get the graph of WikiWikiWeb working on Obsidian.
  - Tried it on a computer with the following specs:
    - Cores:  8
    - CPUs:   16
    - Model:  Intel(R) Core(TM) i7-10700 CPU @ 2.90GHz
    - MHz:    4800.0 	Quadro T1000 Mobile
    - Memory: 31 GB
    - Swap:   16777212 kB
