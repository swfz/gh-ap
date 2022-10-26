package main

import (
	"fmt"
	"log"
	// "encoding/json"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

func main() {

	// sample project
	projectId := "PVT_kwHOAD6fZs4AHaJ0"
	// contentId := "MDU6SXNzdWU5NDEyMDEyNDM="
	itemId := "PVTI_lAHOAD6fZs4AHaJ0zgDB2To"
	// Date Month
	// fieldId := "PVTF_lAHOAD6fZs4AHaJ0zgERRzM"
	// month, _ := time.Parse("2006-01-02", "2022-10-01")
	// month := time.Date(2022, 10, 1, 0, 0, 0, 0, time.Local)
	fieldId := "PVTIF_lAHOAD6fZs4AHaJ0zgERRyI"
	iterationId := "51e3929f"

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {iterationId: $value}})"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId": graphql.ID(itemId),
		"fieldId": graphql.ID(fieldId),
		"value": graphql.String(iterationId),
	}
	fmt.Println(variables)

	err = gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mutation)
}