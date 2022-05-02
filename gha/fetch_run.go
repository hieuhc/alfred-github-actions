package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

var (
	maxAge =  1 * time.Minute
	ghaIconPath string
)

const (
	successIconPath  =  "icons/green-check.png"
	failIconPath  =  "icons/red-fail.png"
	runningIconPath  = "icons/gha-run.png"
)

type RunItem struct {
	Title string
	SubTitle string
	HTMLURL string
	UID string
	IconPath string
	WorkflowName string
	RunNumber string

}

func fetchRun(client *github.Client, context context.Context, owner string, repoName string) ([]RunItem, error) {
	opts := &github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: 50}}
	workflowRuns, _, err := client.Actions.ListWorkflowRunsByFileName(context, owner, repoName, workflow, opts)
	if err != nil {
		return nil, err
	}
	var runItems []RunItem
	workflowName := strings.Split(workflow, ".")[0]
	for _, run := range workflowRuns.WorkflowRuns {
		var status string
		if run.Conclusion != nil {
			status = *run.Conclusion
			if status == "success" {
				ghaIconPath = successIconPath
			} else{
				ghaIconPath = failIconPath
			}

		} else {
			status = "running"
			ghaIconPath = runningIconPath
		}

		createdAt := *run.CreatedAt
		diffMins := int(time.Since(createdAt.Time).Minutes())
		var diffString string = ""
		var diffHour int = diffMins / 60
		if diffHour != 0 {
			diffString += strconv.Itoa(diffHour) + "h"
		}
		diffString += strconv.Itoa(diffMins % 60) + "m"
		branch := *run.HeadBranch
		var commitAuthor string = *run.HeadCommit.Author.Name
		subtitle := *run.Name + " #" + strconv.Itoa(*run.RunNumber) + ": " + diffString + " ago" + " by " + commitAuthor
		title := branch
		runItems = append(runItems, RunItem{
			Title: title,
			SubTitle: subtitle,
			HTMLURL: *run.HTMLURL,
			UID: strconv.Itoa(int(*run.ID)),
			WorkflowName: workflowName,
			RunNumber: strconv.Itoa(*run.RunNumber),
			IconPath: ghaIconPath,
		})
	}
	return runItems, nil
}

func runFetchRun(){
	logger.Println("Start fetch gha workflow run")
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
	runCacheName := fmt.Sprintf("%s_%s_run_%s.json", owner, repoName, workflow)

	reload := func() (interface{}, error) { return fetchRun(client, ctx, owner, repoName) }
	var runItems []RunItem
	// Used only in background job triggered by fetch_workflow
	if cache {
		if err := wf.Cache.LoadOrStoreJSON(runCacheName, maxAge, reload, &runItems); err != nil {
			wf.Fatal(err.Error())
		}
		return
	}
	// Check if the background job started in fetch_workflow is still running
	backgroundJobName := "cache_runs" + owner + repoName + workflow
	for {
		if wf.IsRunning(backgroundJobName){
			logger.Printf("Background job %s is still running", backgroundJobName)
			time.Sleep(200 * time.Millisecond)
		} else {
			break
		}
	}

	if err := wf.Cache.LoadOrStoreJSON(runCacheName, maxAge, reload, &runItems); err != nil {
		wf.Fatal(err.Error())
	}
	for _, item := range runItems {
		ghaRunIcon := aw.Icon{Value: item.IconPath}
		wf.NewItem(item.Title).Subtitle(item.SubTitle).UID(item.UID).Icon(&ghaRunIcon).Arg(item.HTMLURL).Valid(true).NewModifier("cmd").Var("runID", item.UID).Var("runNumber", item.RunNumber).Var("branch", item.Title).Var("workflow", item.WorkflowName)
	}

	if len(query) > 0 {
		logger.Println("query: ", query)
		wf.Filter(query)
	}
	wf.SendFeedback()
}
