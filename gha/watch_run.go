package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)


func checkRun(client *github.Client, context context.Context, owner string, repoName string, runID int64) (*string, error) {
	run, _, err := client.Actions.GetWorkflowRunByID(context, owner, repoName, runID)
	if err != nil {
		return nil, err
	}
	var status string
	if run.Conclusion != nil {
		status = *run.Conclusion
	} else {
		status = "running"
	}
	return &status, nil
}

func runWatchRun(){
	logger.Println("Start watch gha workflow run")
	ctx := context.Background()

	// get token from keychain
	token, err := getToken()
	if err != nil {
		wf.FatalError(err)
	}
	logger.Println("Found Github PAT in keychain")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	repoInfo := strings.Split(repo, "/")
	owner := repoInfo[0]
	repoName := repoInfo[1]

	runID, err := strconv.ParseInt(runIdString, 10, 64) 
	if err != nil {
		wf.Fatal(err.Error())
	}
	var runStatus *string
	for {
		runStatus, err = checkRun(client, ctx, owner, repoName, runID)
		if err != nil {
			wf.Fatal(err.Error())
		}
		logger.Printf("Run status of %s/%s #%s is %s", owner, repoName, runIdString, *runStatus)
		if *runStatus != "running"{
			break
		}
		time.Sleep(10 * time.Second)
	}
	if *runStatus == "failure"{
		os.Exit(1)
	}else if *runStatus == "success"{
		os.Exit(0)
	}
}

