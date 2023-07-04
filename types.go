package main

import "github.com/shurcooL/githubv4"

type Project struct {
	Title string
	Id    string
	Type  string
}

type Content struct {
	Id     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type Option struct {
	Id   string
	Name string
}

type ProjectField struct {
	Id       string
	Name     string
	DataType string
	Options  []Option
}

type Repository struct {
	Name  string `json:"name"`
	Owner struct {
		Id    string `json:"id"`
		Login string `json:"login"`
	} `json:"owner"`
	IsInOrganization bool `json:"isInOrganization"`
}
type NamedDateValue struct {
	Date githubv4.Date `json:"date,omitempty"`
}
