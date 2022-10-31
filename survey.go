package main

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"log"
	"strconv"
	"strings"
	"time"
)

func askOneProjectId(projects []Project) (projectId string) {
	projectIds := make([]string, len(projects))
	for i, node := range projects {
		projectIds[i] = node.Id
		if node.Id == "" {
			log.Print(`[Warning] This extension requires permission for the "project" scope. You may not currently have permission to retrieve project information, please check`)
		}
	}

	qs := []*survey.Question{
		{
			Name: "ProjectId",
			Prompt: &survey.Select{
				Message: "Choose a Project",
				Options: projectIds,
				Description: func(value string, index int) string {
					return projects[index].Title + " (" + projects[index].Type + ")"
				},
				Filter: func(filterValue string, optValue string, optIndex int) bool {
					return strings.Contains(projects[optIndex].Title, filterValue)
				},
			},
		},
	}
	answers := struct{ ProjectId string }{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err.Error())
	}

	return answers.ProjectId
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
	var numbers = make([]string, len(contents))
	for i, c := range contents {
		numbers[i] = strconv.Itoa(c.Number)
	}

	name := contentType + " Number"
	qs := []*survey.Question{
		{
			Name: "number",
			Prompt: &survey.Select{
				Message: name,
				Options: numbers,
				Description: func(value string, index int) string {
					return contents[index].Title
				},
				Filter: func(filterValue string, optValue string, optIndex int) bool {
					return strings.Contains(contents[optIndex].Title, filterValue)
				},
				PageSize: 50,
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
			Name:   fieldName,
			Prompt: &survey.Input{Message: fieldName},
			Validate: func(v interface{}) error {
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
			},
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
			Name:   fieldName,
			Prompt: &survey.Input{Message: fieldName},
			Validate: func(v interface{}) error {
				strValue := v.(string)
				_, err := strconv.ParseFloat(strValue, 64)
				if err != nil {
					return errors.New("Value is Int or Float.")
				}
				return nil
			},
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
	fieldOptionSize := len(options)
	optionNames := make([]string, fieldOptionSize)
	for i, opt := range options {
		optionNames[i] = opt.Name
	}
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
