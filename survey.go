package main

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"log"
	"strconv"
	"strings"
	"time"
)

// projectOptionNames はプロジェクト選択肢の表示名と、Id を取得できなかったプロジェクトの数を返す
func projectOptionNames(projects []Project) (names []string, missingIdCount int) {
	names = make([]string, len(projects))
	for i, node := range projects {
		names[i] = node.Title
		if node.Id == "" {
			missingIdCount++
		}
	}
	return names, missingIdCount
}

func projectDescription(projects []Project) func(value string, index int) string {
	return func(value string, index int) string {
		return projects[index].Type
	}
}

func contentNumberOptions(contents []Content) []string {
	numbers := make([]string, len(contents))
	for i, c := range contents {
		numbers[i] = strconv.Itoa(c.Number)
	}
	return numbers
}

func contentDescription(contents []Content) func(value string, index int) string {
	return func(value string, index int) string {
		return contents[index].Title
	}
}

// contentFilter は入力値をタイトルに部分一致で含むコンテンツだけを候補に残す
func contentFilter(contents []Content) func(filterValue string, optValue string, optIndex int) bool {
	return func(filterValue string, optValue string, optIndex int) bool {
		return strings.Contains(contents[optIndex].Title, filterValue)
	}
}

func selectOptionNames(options []Option) []string {
	names := make([]string, len(options))
	for i, opt := range options {
		names[i] = opt.Name
	}
	return names
}

// validateDate は YYYY-MM-DD 形式か空文字(未入力)を受け付ける
func validateDate(v interface{}) error {
	strValue := v.(string)
	// Allow Zero Value
	if strValue == "" {
		return nil
	}
	_, err := time.Parse("2006-01-02", strValue)
	if err != nil {
		return errors.New("Please format it like this '2006-01-02'.")
	}
	return nil
}

func validateNumber(v interface{}) error {
	strValue := v.(string)
	_, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return errors.New("Value is Int or Float.")
	}
	return nil
}

func askOneProjectId(projects []Project) (projectId string) {
	projectNames, missingIdCount := projectOptionNames(projects)
	for i := 0; i < missingIdCount; i++ {
		log.Print(`[Warning] This extension requires permission for the "project" scope. You may not currently have permission to retrieve project information, please check`)
	}

	qs := []*survey.Question{
		{
			Name: "ProjectId",
			Prompt: &survey.Select{
				Message:     "Choose a Project",
				Options:     projectNames,
				Description: projectDescription(projects),
			},
		},
	}
	answers := map[string]interface{}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}

	optionAnswer := answers["ProjectId"].(survey.OptionAnswer)

	return projects[optionAnswer.Index].Id
}

func askOneContentType(itemTypes []string) string {
	var selectedType string

	typeQuestion := &survey.Select{
		Message: "Choose a Item Type",
		Options: itemTypes,
		Default: itemTypes[0],
	}
	survey.AskOne(typeQuestion, &selectedType)

	return selectedType
}

func askContentNumber(contentType string, contents []Content) string {
	numbers := contentNumberOptions(contents)

	name := contentType + " Number"
	qs := []*survey.Question{
		{
			Name: "number",
			Prompt: &survey.Select{
				Message:     name,
				Options:     numbers,
				Description: contentDescription(contents),
				Filter:      contentFilter(contents),
				PageSize:    50,
			},
		},
	}
	answers := map[string]interface{}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}
	optionAnswer := answers["number"].(survey.OptionAnswer)

	return optionAnswer.Value
}

func askTextFieldValue(fieldName string) string {
	input := ""
	prompt := &survey.Input{
		Message: fieldName,
	}
	survey.AskOne(prompt, &input)

	return input
}

func askDateFieldValue(fieldName string) string {
	qs := []*survey.Question{
		{
			Name:     fieldName,
			Prompt:   &survey.Input{Message: fieldName},
			Validate: validateDate,
		},
	}
	answers := map[string]interface{}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}

	return answers[fieldName].(string)
}

func askNumberFieldValue(fieldName string) float64 {
	qs := []*survey.Question{
		{
			Name:     fieldName,
			Prompt:   &survey.Input{Message: fieldName},
			Validate: validateNumber,
		},
	}

	answers := map[string]interface{}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}
	f, _ := strconv.ParseFloat(answers[fieldName].(string), 64)

	return f
}

func askOneSelectFieldValue(fieldName string, options []Option) string {
	optionNames := selectOptionNames(options)
	qs := []*survey.Question{
		{
			Name: fieldName,
			Prompt: &survey.Select{
				Message: fieldName,
				Options: optionNames,
			},
		},
	}
	answers := map[string]interface{}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}
	optionAnswer := answers[fieldName].(survey.OptionAnswer)

	return options[optionAnswer.Index].Id
}
