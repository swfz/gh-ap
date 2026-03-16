package main

import (
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"log"
)

func queryUserProjects(gqlclient api.GQLClient, login string) (projects []struct {
	Title string
	Id    string
}) {
	var query struct {
		User struct {
			ProjectsV2 struct {
				Nodes []struct {
					Title string
					Id    string
				}
			} `graphql:"projectsV2(first: $size)"`
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login": graphql.String(login),
		"size":  graphql.Int(10),
	}

	err := gqlclient.Query("ProjectsV2", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.User.ProjectsV2.Nodes
}

func queryOrganizationProjects(gqlclient api.GQLClient, owner string) (projects []struct {
	Title string
	Id    string
}) {
	var query struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []struct {
					Title string
					Id    string
				}
			} `graphql:"projectsV2(first: $size)"`
		} `graphql:"organization(login: $login)"`
	}
	variables := map[string]interface{}{
		"login": graphql.String(owner),
		"size":  graphql.Int(10),
	}

	err := gqlclient.Query("ProjectsV2", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.Organization.ProjectsV2.Nodes
}

func queryProjectFieldTypes(gqlclient api.GQLClient, projectId string) []FieldType {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						ProjectV2FieldCommon struct {
							Id       string
							Name     string
							DataType string
						} `graphql:"... on ProjectV2FieldCommon"`
					} `graphql:"nodes"`
				} `graphql:"fields(first: $number)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"number":    graphql.Int(20),
	}

	err := gqlclient.Query("FieldTypes", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	fieldTypes := make([]FieldType, len(query.Node.ProjectV2.Fields.Nodes))
	for i, node := range query.Node.ProjectV2.Fields.Nodes {
		fieldTypes[i] = FieldType{
			Id:       node.ProjectV2FieldCommon.Id,
			Name:     node.ProjectV2FieldCommon.Name,
			DataType: node.ProjectV2FieldCommon.DataType,
		}
	}

	return fieldTypes
}

func queryProjectField(gqlclient api.GQLClient, projectId string) []ProjectFieldNode {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []ProjectFieldNode `graphql:"nodes"`
				} `graphql:"fields(first: $number)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectId),
		"number":    graphql.Int(20),
	}

	err := gqlclient.Query("Fields", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.Node.ProjectV2.Fields.Nodes
}
