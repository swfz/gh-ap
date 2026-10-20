package main

import (
	"encoding/json"
	gh "github.com/cli/go-gh/v2"
	"log"
)

// テストで差し替えるためのパッケージ変数
var (
	ghExec   = gh.Exec
	logFatal = log.Fatal
)

func ghRepository() Repository {
	args := []string{"repo", "view", "--json", "name,owner,isInOrganization"}
	stdOut, _, err := ghExec(args...)
	if err != nil {
		logFatal(err)
	}
	var repository Repository
	if err := json.Unmarshal(stdOut.Bytes(), &repository); err != nil {
		panic(err)
	}

	return repository
}

func ghCurrentPullRequest() Content {
	args := []string{"pr", "view", "--json", "id,number,title"}
	stdOut, _, err := ghExec(args...)
	if err != nil {
		logFatal(err)
	}

	var currentPR Content
	if err := json.Unmarshal(stdOut.Bytes(), &currentPR); err != nil {
		panic(err)
	}

	return currentPR
}

func ghContent(contentType string, number string) Content {
	var subCommand string
	if contentType == "Issue" {
		subCommand = "issue"
	} else {
		subCommand = "pr"
	}
	args := []string{subCommand, "view", number, "--json", "id,number,title"}
	stdOut, _, err := ghExec(args...)
	if err != nil {
		logFatal("Error: not found " + contentType)
	}
	var content Content
	if err := json.Unmarshal(stdOut.Bytes(), &content); err != nil {
		panic(err)
	}

	return content
}

func ghContentList(contentType string) []Content {
	var subCommand string
	if contentType == "Issue" {
		subCommand = "issue"
	} else {
		subCommand = "pr"
	}
	args := []string{subCommand, "list", "--limit", "50", "--json", "id,number,title"}
	stdOut, _, err := ghExec(args...)
	if err != nil {
		logFatal(err)
	}
	var contents []Content

	if err := json.Unmarshal(stdOut.Bytes(), &contents); err != nil {
		panic(err)
	}

	return contents
}
