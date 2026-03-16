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
	// Single Select Status
	fieldId := "PVTSSF_lAHOAD6fZs4AHaJ0zgEQ3Vs"
	statusId := "47fc9ee4"

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId itemId: $itemId fieldId: $fieldId value: {singleSelectOptionId: $value}})"`
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
		"value": graphql.String(statusId),
	}
	fmt.Println(variables)

	err = gqlclient.Mutate("UpdateFieldValue", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mutation)
}