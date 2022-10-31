package main

import (
	"flag"
	"fmt"
)

func main() {
	var options struct {
		issueNo int
		prNo    int
	}

	flag.IntVar(&options.issueNo, "issue", 0, "Issue Number")
	flag.IntVar(&options.prNo, "pr", 0, "PullRequest Number")
	flag.Parse()

	fmt.Printf("%+v\n", options)
}
