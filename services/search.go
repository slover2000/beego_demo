package services

import (
	"fmt"
	"golang.org/x/net/context"

	"github.com/olivere/elastic"
)

func SearchUser(id string) {
	client, err := elastic.NewClient(elastic.SetURL("http://192.168.2.10:9201"))
	if err != nil {
		// Handle error
	}

	get, err := client.Get().Index("users").Type("doc").Id(id).Do(context.Background())
	if err != nil {
		// Handle error
		panic(err)
	}

	if get.Found {
		fmt.Printf("Got document %s in version %d from index %s, type %s\n", get.Id, get.Version, get.Index, get.Type)
	}	
}

func SearchUserByName(name string) {
	client, err := elastic.NewClient(elastic.SetURL("http://192.168.2.10:9201", "http://127.0.0.1:9201"))
	if err != nil {
		// Handle error
	}	
	// Search with a term query
	termQuery := elastic.NewTermQuery("name", name)
	searchResult, err := client.Search().
		Index("users").        // search in index "users"
		Query(termQuery).        // specify the query
		Sort("user", true).      // sort by "user" field, ascending
		From(0).Size(10).        // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
}