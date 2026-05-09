package main

import (
	"fmt"
	"log"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

func calc(v1 int, v2 int) (sum int, mul int) {
	sum = v1 + v2
	mul = v1 * v2

	return
}

func main() {
	fmt.Println("hi world, this is the gh-ap extension!!!")
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	response := struct {Login string}{}
	err = client.Get("user", &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	sum, mul := calc(1,2)
	fmt.Printf("%d, %d", sum, mul)

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var query struct {
		User struct {
			ProjectsV2 struct {
				Nodes []struct {
					Title string
					Id string
				}
			} `graphql:"projectsV2(first: $projects)"`
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]interface{}{
		"login": graphql.String(response.Login),
		"projects": graphql.Int(10),
	}

	err = gqlclient.Query("ProjectsV2", &query, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(query)
	fmt.Println(variables)

	fmt.Printf("running as %s\n", response.Login)
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
