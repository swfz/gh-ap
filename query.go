package main

import (
	graphql "github.com/cli/shurcooL-graphql"
	"log"
)

func queryUserProjects(gqlclient GQLClient, login string) (projects []struct {
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

func queryOrganizationProjects(gqlclient GQLClient, owner string) (projects []struct {
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

func queryUserProjectByNumber(gqlclient GQLClient, login string, number int) (string, bool) {
	var query struct {
		User struct {
			ProjectV2 struct {
				Id string
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login":  graphql.String(login),
		"number": graphql.Int(number),
	}

	err := gqlclient.Query("UserProjectV2", &query, variables)
	if err != nil {
		return "", false
	}

	if query.User.ProjectV2.Id == "" {
		return "", false
	}

	return query.User.ProjectV2.Id, true
}

func queryOrganizationProjectByNumber(gqlclient GQLClient, owner string, number int) (string, bool) {
	var query struct {
		Organization struct {
			ProjectV2 struct {
				Id string
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $login)"`
	}
	variables := map[string]interface{}{
		"login":  graphql.String(owner),
		"number": graphql.Int(number),
	}

	err := gqlclient.Query("OrgProjectV2", &query, variables)
	if err != nil {
		return "", false
	}

	if query.Organization.ProjectV2.Id == "" {
		return "", false
	}

	return query.Organization.ProjectV2.Id, true
}

func queryProjectFieldTypes(gqlclient GQLClient, projectId string) []FieldType {
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

func queryProjectField(gqlclient GQLClient, projectId string) []ProjectFieldNode {
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
