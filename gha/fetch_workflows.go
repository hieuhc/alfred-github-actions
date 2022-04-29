package main

import (
	"context"
	"flag"
	"os/exec"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

// Workflow is the main API
var (
	maxCacheAge =  10 * time.Minute
	workflowIcon  = &aw.Icon{Value: "icons/gha_wf.png"}
)

type GHAWorkflow struct {
	Name string
	FileName string
	UID string
	HTMLURL string
}

func runFetchWorkflow(){
	logger.Println("Start fetch gha workflows alfred workflow")
	flag.Parse()
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
	cachedWorkflowsName := owner + "_" + repoName + "_workflows.json"
	if cache {
		// logger.Printf("Caching workflows for %s",repoInfo)
		var workflowToCache []GHAWorkflow
		workflowStruct, _, _ := client.Actions.ListWorkflows(ctx, owner, repoName, nil)
		for _, ghaWf := range workflowStruct.Workflows {
			fileName := strings.Split(*ghaWf.Path, "/")[2]
			workflowToCache = append(workflowToCache, GHAWorkflow{
				Name: *ghaWf.Name,
				FileName: fileName,
				UID: strconv.Itoa(int(*ghaWf.ID)),
				HTMLURL: strings.Replace(*ghaWf.HTMLURL, "blob/master/.github", "actions", 1),
			})
		}
		err := wf.Cache.StoreJSON(cachedWorkflowsName, workflowToCache)
		if err != nil {
			wf.Fatal("Can not fetch workflows for caching")
		}
		return
	}

	var GhaWorkflows []GHAWorkflow
	if wf.Cache.Exists(cachedWorkflowsName) {
		logger.Printf("Loading cached workflows")
		if err := wf.Cache.LoadJSON(cachedWorkflowsName, &GhaWorkflows); err != nil {
			wf.Fatal(err.Error())
		}
	}
	if wf.Cache.Expired(cachedWorkflowsName, maxCacheAge) {
		logger.Printf("Cache is expired or not existed !!")
		wf.Rerun(0.5)
		if !wf.IsRunning("cachingWorkflows") {
			cmd := exec.Command("./bin/main", "-stage", "fetch_workflow", "-cache", "-repo", repo)
			logger.Printf("Run in background with comand %s", cmd.String())
			if err := wf.RunInBackground("cachingWorkflows", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			logger.Printf("Backround job is already running")
		}
		if len(GhaWorkflows) == 0 {
			wf.NewItem("Caching workflows...").Icon(aw.IconInfo)
			wf.SendFeedback()
			return
		}
	}
	for _, ghaWf := range GhaWorkflows {
		cmd := exec.Command("./bin/main", "-stage", "fetch_run", "-cache", "-repo", repo, "-workflow", ghaWf.FileName)
		logger.Printf("Run in background with comand %s", cmd.String())

		jobName := "cache_runs" + owner + repoName + ghaWf.FileName
		wf.RunInBackground(jobName, cmd)
		wf.NewItem(ghaWf.Name).Arg(ghaWf.FileName).Subtitle("").UID(ghaWf.UID).Icon(workflowIcon).Valid(true).NewModifier("cmd").Arg(ghaWf.HTMLURL)
	}
	if len(query) > 0 {
		logger.Println("query: ", query)
		wf.Filter(query)
	}
	wf.SendFeedback()
}



