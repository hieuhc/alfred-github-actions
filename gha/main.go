package main

import (
	"flag"
	"log"
	"os"

	aw "github.com/deanishe/awgo"
)
const (
	cachedGithubRepos = "cached-github-repositories.json"
	keychainService = "alfred_gha_token"
)


// Workflow is the main API
var (
	logger = log.New(os.Stderr, "logger", log.LstdFlags)
	wf *aw.Workflow
	stage string
	repo string
	workflow string
	runIdString string

	query string

	cache bool
	// entry run
	ghToken string
)

func init(){
	wf = aw.New()
	flag.StringVar(&stage, "stage", "", "stage that triggers this main file")
	// entry run
	flag.StringVar(&ghToken, "ghToken", "", "token for github authentication")

	flag.StringVar(&query, "query", "", "query string for results filtering")

	flag.StringVar(&repo, "repo", "", "github repository to fetch workflows")
	flag.StringVar(&workflow, "workflow", "", "workflow ID to fetch runs")
	flag.StringVar(&runIdString, "runID", "", "runID to poll for status")

	flag.BoolVar(&cache, "cache", false, "whether the repo's workflows is cached'")

}

func runMain(){
	logger.Println("Start main run")
	flag.Parse()
	// ctx := context.Background()
	if stage == "entry" {
		runEntry()
	} else if stage == "fetch_repo" {
		runFetchRepo()
	} else if stage == "fetch_workflow" {
		runFetchWorkflow()
	} else if stage == "fetch_run" {
		runFetchRun()
	} else if stage == "watch_run" {
		runWatchRun()
	} 
}


func main(){
	wf.Run(runMain)
}


