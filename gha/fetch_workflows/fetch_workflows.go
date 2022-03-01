package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v41/github"
	"github.com/keybase/go-keychain"
	"golang.org/x/oauth2"
)

// Workflow is the main API
var (
	logger = log.New(os.Stderr, "logger", log.LstdFlags)
	maxCacheAge =  10 * time.Minute
	workflowIcon  = &aw.Icon{Value: "icons/gha_wf.png"}
	wf *aw.Workflow
	repo string
	query string
	cache bool
)

type GHAWorkflow struct {
	Name string
	FileName string
	UID string
	HTMLURL string
}

func init(){
	wf = aw.New()
	flag.StringVar(&repo, "repo", "", "github repository to fetch workflows")
	flag.StringVar(&query, "query", "", "gha workflow to fetch instances in next step")
	flag.BoolVar(&cache, "cache", false, "whether the repo's workflows is cached'")

}

func run(){
	logger.Println("Start fetch gha workflows alfred workflow")
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
				cmd := exec.Command("./bin/fetch_workflows", "-cache", "-repo", repo)
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
			cmd := exec.Command("./bin/fetch_runs", "-cache", "-repo", repo, "-workflow", ghaWf.FileName)
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
}


func main(){
	wf.Run(run)
}


