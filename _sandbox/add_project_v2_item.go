package main

import (
	"fmt"
	"log"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

func main() {

	// sample project
	projectId := "PVT_kwHOAD6fZs4AHaJ0"
	contentId := "MDU6SXNzdWU5NDEyMDEyNDM="

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				Id string
			}
		} `graphql:"addProjectV2ItemById(input: {contentId: $contentId projectId: $projectId})"`
	}

	variables := map[string]interface{}{
		"contentId": graphql.ID(contentId),
		"projectId": graphql.ID(projectId),
	}

	err = gqlclient.Mutate("Assign", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mutation)
}