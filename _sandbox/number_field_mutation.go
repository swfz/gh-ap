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
	// Number Point
	fieldId := "PVTF_lAHOAD6fZs4AHaJ0zgERRxE"

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
			// Item struct {
				// Id string
			// }
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {number: $value}})"`
	}

	// numberValue := map[string]interface{}{
	// 	"number": graphql.Int(3),
	// }
	// v,err := json.Marshal(numberValue)
	// fmt.Println(string(v))
	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"itemId": graphql.ID(itemId),
		"fieldId": graphql.ID(fieldId),
		"value": graphql.Float(3),
	}
	fmt.Println(variables)

	err = gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mutation)
}