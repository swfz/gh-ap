package main

import (
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"log"
)

func queryProjects(gqlclient api.GQLClient, login string) (projects []struct {
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
			} `graphql:"projectsV2(first: $projects)"`
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login":    graphql.String(login),
		"projects": graphql.Int(10),
	}

	err := gqlclient.Query("ProjectsV2", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.User.ProjectsV2.Nodes
}

func queryProjectFieldTypes(gqlclient api.GQLClient, projectId string) (fieldTypes []struct {
	Id       string
	Name     string
	DataType string
}) {
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

	nodes := len(query.Node.ProjectV2.Fields.Nodes)
	fieldTypes = make([]struct {
		Id       string
		Name     string
		DataType string
	}, nodes)

	for i, node := range query.Node.ProjectV2.Fields.Nodes {
		fieldTypes[i] = node.ProjectV2FieldCommon
	}

	return fieldTypes
}

func queryProjectField(gqlclient api.GQLClient, projectId string) []struct {
	ProjectV2IterationField struct {
		Id            string
		Name          string
		Configuration struct {
			Iterations []struct {
				StartDate string
				Id        string
			}
		}
	} `graphql:"... on ProjectV2IterationField"`
	ProjectV2SingleSelectField struct {
		Id      string
		Name    string
		Options []struct {
			Id   string
			Name string
		}
	} `graphql:"... on ProjectV2SingleSelectField"`
} {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						ProjectV2IterationField struct {
							Id            string
							Name          string
							Configuration struct {
								Iterations []struct {
									StartDate string
									Id        string
								}
							}
						} `graphql:"... on ProjectV2IterationField"`
						ProjectV2SingleSelectField struct {
							Id      string
							Name    string
							Options []struct {
								Id   string
								Name string
							}
						} `graphql:"... on ProjectV2SingleSelectField"`
					} `graphql:"nodes"`
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
