package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cli/go-gh/v2/pkg/api"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type RESTClient interface {
	Get(path string, response interface{}) error
}

func getProjects(gqlclient GQLClient, restClient RESTClient, getRepository func() Repository) []Project {
	response := struct{ Login string }{}
	err := restClient.Get("user", &response)
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

	repository := getRepository()

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

func resolveProjectIdByNumber(gqlclient GQLClient, restClient RESTClient, getRepository func() Repository, number int) (string, bool) {
	response := struct{ Login string }{}
	err := restClient.Get("user", &response)
	if err != nil {
		log.Fatal(err)
	}

	// Try user project first
	if id, found := queryUserProjectByNumber(gqlclient, response.Login, number); found {
		return id, true
	}

	// Try organization project
	repository := getRepository()
	if repository.IsInOrganization {
		if id, found := queryOrganizationProjectByNumber(gqlclient, repository.Owner.Login, number); found {
			return id, true
		}
	}

	return "", false
}

func findProjectByName(projects []Project, name string) (string, bool) {
	for _, p := range projects {
		if p.Title == name {
			return p.Id, true
		}
	}
	return "", false
}

// fieldPrompter は CLI で値が指定されなかったフィールドを対話で尋ねる
type fieldPrompter struct {
	askText   func(fieldName string) string
	askDate   func(fieldName string) string
	askNumber func(fieldName string) float64
	askSelect func(fieldName string, options []Option) string
}

var surveyPrompter = fieldPrompter{
	askText:   askTextFieldValue,
	askDate:   askDateFieldValue,
	askNumber: askNumberFieldValue,
	askSelect: askOneSelectFieldValue,
}

type invalidNumberError struct {
	field string
	value string
}

func (e invalidNumberError) Error() string {
	return fmt.Sprintf("Invalid number value for field %s: %s", e.field, e.value)
}

type optionNotFoundError struct {
	field string
	value string
}

func (e optionNotFoundError) Error() string {
	return fmt.Sprintf("Option '%s' not found for field '%s'", e.value, e.field)
}

func applyFieldValues(gqlclient GQLClient, projectId string, itemId string, fields []ProjectField, fieldValues map[string]string, prompter fieldPrompter) error {
	for _, field := range fields {
		cliValue, hasCLIValue := fieldValues[field.Name]

		if field.DataType == "TEXT" {
			var input string
			if hasCLIValue {
				input = cliValue
			} else {
				input = prompter.askText(field.Name)
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
				dateInput = prompter.askDate(field.Name)
			}
			if dateInput != "" {
				updateDateProjectField(gqlclient, projectId, itemId, field.Id, dateInput)
			}
		}
		if field.DataType == "NUMBER" {
			if hasCLIValue {
				f, err := strconv.ParseFloat(cliValue, 64)
				if err != nil {
					return invalidNumberError{field: field.Name, value: cliValue}
				}
				updateNumberProjectField(gqlclient, projectId, itemId, field.Id, f)
			} else {
				f := prompter.askNumber(field.Name)
				updateNumberProjectField(gqlclient, projectId, itemId, field.Id, f)
			}
		}
		if field.DataType == "SINGLE_SELECT" || field.DataType == "ITERATION" {
			if hasCLIValue {
				opt, found := findOptionByName(field.Options, cliValue)
				if !found {
					return optionNotFoundError{field: field.Name, value: cliValue}
				}
				if field.DataType == "ITERATION" {
					updateIterationProjectField(gqlclient, projectId, itemId, field.Id, opt.Id)
				} else {
					updateSingleSelectProjectField(gqlclient, projectId, itemId, field.Id, opt.Id)
				}
			} else {
				selected := prompter.askSelect(field.Name, field.Options)
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

	return nil
}

// interactivePrompter は対話で値を尋ねる処理をまとめたもの(テストで差し替える)
type interactivePrompter struct {
	fields           fieldPrompter
	askProject       func(projects []Project) string
	askContentType   func(itemTypes []string) string
	askContentNumber func(contentType string, contents []Content) string
}

var surveyInteractivePrompter = interactivePrompter{
	fields:           surveyPrompter,
	askProject:       askOneProjectId,
	askContentType:   askOneContentType,
	askContentNumber: askContentNumber,
}

type options struct {
	issueNo       int
	prNo          int
	projectName   string
	projectNumber int
	fields        fieldFlags
}

func run(opts options, gqlclient GQLClient, restClient RESTClient, getRepository func() Repository, prompter interactivePrompter, out io.Writer) error {
	if opts.projectName != "" && opts.projectNumber != 0 {
		return errors.New("-project and -project-id cannot be used together")
	}

	var projectId string
	if opts.projectNumber != 0 {
		var found bool
		projectId, found = resolveProjectIdByNumber(gqlclient, restClient, getRepository, opts.projectNumber)
		if !found {
			return fmt.Errorf("Project #%d not found", opts.projectNumber)
		}
	} else {
		projects := getProjects(gqlclient, restClient, getRepository)
		if opts.projectName != "" {
			var found bool
			projectId, found = findProjectByName(projects, opts.projectName)
			if !found {
				return fmt.Errorf("Project '%s' not found", opts.projectName)
			}
		} else {
			projectId = prompter.askProject(projects)
		}
	}
	fields := getProjectFields(gqlclient, projectId)

	var content Content
	if opts.issueNo != 0 {
		content = ghContent("Issue", strconv.Itoa(opts.issueNo))
	} else if opts.prNo != 0 {
		content = ghContent("PullRequest", strconv.Itoa(opts.prNo))
	} else {
		itemTypes := []string{"Current PullRequest", "PullRequest", "Issue"}
		selectedType := prompter.askContentType(itemTypes)

		if selectedType == "Current PullRequest" {
			content = ghCurrentPullRequest()
		} else {
			contentList := ghContentList(selectedType)
			number := prompter.askContentNumber(selectedType, contentList)
			content = ghContent(selectedType, number)
		}
	}
	itemId := addProject(gqlclient, projectId, content.Id)

	fieldValues := parseFieldFlags(opts.fields)
	if err := applyFieldValues(gqlclient, projectId, itemId, fields, fieldValues, prompter.fields); err != nil {
		return err
	}

	printItemSummary(out, buildItemSummary(queryProjectItem(gqlclient, itemId)))
	return nil
}

func main() {
	var opts options
	flag.IntVar(&opts.issueNo, "issue", 0, "Issue Number")
	flag.IntVar(&opts.prNo, "pr", 0, "PullRequest Number")
	flag.StringVar(&opts.projectName, "project", "", "Project Name")
	flag.IntVar(&opts.projectNumber, "project-id", 0, "Project Number (the number shown in the project URL)")
	flag.Var(&opts.fields, "field", "Field value in 'FieldName=Value' format (can be specified multiple times)")
	flag.Parse()

	gqlclient, err := api.DefaultGraphQLClient()
	if err != nil {
		log.Fatal(err)
	}
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(opts, gqlclient, restClient, ghRepository, surveyInteractivePrompter, os.Stdout); err != nil {
		log.Fatal(err)
	}
}
