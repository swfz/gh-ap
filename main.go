package main

import (
	"flag"
	"fmt"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"log"
	"strconv"
	"strings"
)

func getProjects(gqlclient api.GQLClient) []Project {
	restClient, err := gh.RESTClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	response := struct{ Login string }{}
	err = restClient.Get("user", &response)
	if err != nil {
		log.Fatal(err)
	}
	var projects []Project

	userProjects := queryUserProjects(gqlclient, response.Login)
	for _, p := range userProjects {
		project := struct {
			Title string
			Id    string
			Type  string
		}{
			Title: p.Title,
			Id:    p.Id,
			Type:  "UserProject",
		}
		projects = append(projects, project)
	}

	repository := ghRepository()

	if repository.IsInOrganization {
		organizationProjects := queryOrganizationProjects(gqlclient, repository.Owner.Login)
		for _, p := range organizationProjects {
			project := struct {
				Title string
				Id    string
				Type  string
			}{
				Title: p.Title,
				Id:    p.Id,
				Type:  "OrganizationProject",
			}
			projects = append(projects, project)
		}
	}

	return projects
}

type fieldFlags []string

func (f *fieldFlags) String() string {
	return strings.Join(*f, ", ")
}

func (f *fieldFlags) Set(value string) error {
	if !strings.Contains(value, "=") {
		return fmt.Errorf("field format must be 'FieldName=Value'")
	}
	*f = append(*f, value)
	return nil
}

func parseFieldFlags(flags fieldFlags) map[string]string {
	fieldValues := make(map[string]string)
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		fieldValues[parts[0]] = parts[1]
	}
	return fieldValues
}

func findOptionByName(options []Option, name string) (Option, bool) {
	for _, opt := range options {
		if opt.Name == name {
			return opt, true
		}
	}
	return Option{}, false
}

func main() {
	var options struct {
		issueNo int
		prNo    int
		fields  fieldFlags
	}
	flag.IntVar(&options.issueNo, "issue", 0, "Issue Number")
	flag.IntVar(&options.prNo, "pr", 0, "PullRequest Number")
	flag.Var(&options.fields, "field", "Field value in 'FieldName=Value' format (can be specified multiple times)")
	flag.Parse()

	gqlclient, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	projects := getProjects(gqlclient)
	projectId := askOneProjectId(projects)
	fields := getProjectFields(gqlclient, projectId)

	var itemId string
	var content Content

	if options.issueNo != 0 || options.prNo != 0 {
		if options.issueNo != 0 {
			content = ghContent("Issue", strconv.Itoa(options.issueNo))
		} else {
			content = ghContent("PullRequest", strconv.Itoa(options.prNo))
		}
	} else {
		itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}
		selectedType := askOneContentType(itemTypes)

		if selectedType == "Current PullRequest" {
			content = ghCurrentPullRequest()
		} else {
			contentList := ghContentList(selectedType)
			number := askContentNumber(selectedType, contentList)
			content = ghContent(selectedType, number)
		}
	}
	itemId = addProject(gqlclient, projectId, content.Id)

	fieldValues := parseFieldFlags(options.fields)

	for _, field := range fields {
		cliValue, hasCLIValue := fieldValues[field.Name]

		if field.DataType == "TEXT" {
			var input string
			if hasCLIValue {
				input = cliValue
			} else {
				input = askTextFieldValue(field.Name)
			}
			if input != "" {
				updateTextProjectField(gqlclient, projectId, itemId, field.Id, input)
			}
		}
		if field.DataType == "DATE" {
			var dateInput string
			if hasCLIValue {
				dateInput = cliValue
			} else {
				dateInput = askDateFieldValue(field.Name)
			}
			if dateInput != "" {
				updateDateProjectField(gqlclient, projectId, itemId, field.Id, dateInput)
			}
		}
		if field.DataType == "NUMBER" {
			if hasCLIValue {
				f, err := strconv.ParseFloat(cliValue, 64)
				if err != nil {
					log.Fatalf("Invalid number value for field %s: %s", field.Name, cliValue)
				}
				updateNumberProjectField(gqlclient, projectId, itemId, field.Id, f)
			} else {
				f := askNumberFieldValue(field.Name)
				updateNumberProjectField(gqlclient, projectId, itemId, field.Id, f)
			}
		}
		if field.DataType == "SINGLE_SELECT" || field.DataType == "ITERATION" {
			if hasCLIValue {
				opt, found := findOptionByName(field.Options, cliValue)
				if !found {
					log.Fatalf("Option '%s' not found for field '%s'", cliValue, field.Name)
				}
				if field.DataType == "ITERATION" {
					updateIterationProjectField(gqlclient, projectId, itemId, field.Id, opt.Id)
				} else {
					updateSingleSelectProjectField(gqlclient, projectId, itemId, field.Id, opt.Id)
				}
			} else {
				selected := askOneSelectFieldValue(field.Name, field.Options)
				if selected == "Skip" {
					continue
				}
				if field.DataType == "ITERATION" {
					updateIterationProjectField(gqlclient, projectId, itemId, field.Id, selected)
				} else {
					updateSingleSelectProjectField(gqlclient, projectId, itemId, field.Id, selected)
				}
			}
		}
	}
}
