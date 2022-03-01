package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v41/github"
	"github.com/keybase/go-keychain"
	"go.deanishe.net/fuzzy"
	"golang.org/x/oauth2"
)

// Workflow is the main API
var (
	logger = log.New(os.Stderr, "logger", log.LstdFlags)
	maxAge =  3 * time.Minute
	wf *aw.Workflow
	repo string
	query string
	workflow string
	cache bool
	ghaIconPath string
)

const (
	successIconPath  =  "icons/green-check.png"
	failIconPath  =  "icons/red-fail.png"
	runningIconPath  = "icons/gha-run.png"
)

func init(){
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.SortOptions(sopts...))
	flag.StringVar(&repo, "repo", "", "github repository to fetch workflows")
	flag.StringVar(&workflow, "workflow", "", "workflow ID to fetch runs")
	flag.StringVar(&query, "query", "", "gha workflow to fetch instances in next step")
	flag.BoolVar(&cache, "cache", false, "cache runs from a workflow")

}
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

func run(){
	logger.Println("Start fetch gha workflow run")
	flag.Parse()
	ctx := context.Background()

	// get token from keychain
	keychain_service := "alfred_gha"
	token := keychain.NewItem()
	token.SetSecClass(keychain.SecClassGenericPassword)
	token.SetService(keychain_service)
	token.SetMatchLimit(keychain.MatchLimitOne)
	token.SetReturnData(true)
	results, err := keychain.QueryItem(token)

	if err != nil {
		logger.Println("Error", err)
		wf.Fatal(err.Error())
	} else if len(results) != 1 {
		logger.Println("Github PAT not found in keychain")
		return
	} else {
		token := string(results[0].Data)
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
		if cache {
			if err := wf.Cache.LoadOrStoreJSON(runCacheName, maxAge, reload, &runItems); err != nil {
				wf.Fatal(err.Error())
			}
			return
		}
		// Check if the background job started at workflow level is still running
		// TODO sync with background jobname in fetch_workflows
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
			// vars := wf.Var([string]string{"runID": item.UID, "branch": item.Title, "workflow": workflow, "runURL": item.HTMLURL})
			wf.NewItem(item.Title).Subtitle(item.SubTitle).UID(item.UID).Icon(&ghaRunIcon).Arg(item.HTMLURL).Valid(true).NewModifier("cmd").Var("runID", item.UID).Var("runNumber", item.RunNumber).Var("branch", item.Title).Var("workflow", item.WorkflowName)
		}

		if len(query) > 0 {
			logger.Println("query: ", query)
			wf.Filter(query)
		}
		wf.SendFeedback()
	}
}


func main(){
	wf.Run(run)
}


