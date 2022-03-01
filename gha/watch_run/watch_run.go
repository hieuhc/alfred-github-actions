package main

import (
	"context"
	"flag"
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
	// TODO fix logger name
	logger = log.New(os.Stderr, "logger", log.LstdFlags)
	wf *aw.Workflow
	repo string
	runIdString string
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
	flag.StringVar(&runIdString, "runID", "", "runID to poll for status")
}

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

func run(){
	logger.Println("Start watch gha workflow run")
	flag.Parse()
	ctx := context.Background()

	// get token from keychain
	// TODO common fetch token
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

}


func main(){
	wf.Run(run)
}


