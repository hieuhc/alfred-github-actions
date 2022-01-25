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
	"golang.org/x/oauth2"
	"go.deanishe.net/fuzzy"
)

// Workflow is the main API
var (
	logger = log.New(os.Stderr, "prefixLogger", log.LstdFlags)
	successIcon  = &aw.Icon{Value: "icons/green-check.png"}
	failIcon  = &aw.Icon{Value: "icons/red-fail.png"}
	runningIcon  = &aw.Icon{Value: "icons/gha-run.png"}
	wf *aw.Workflow
	repo string
	query string
	workflow string
	ghaRunIcon *aw.Icon
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
		opts := &github.ListWorkflowRunsOptions{ListOptions: github.ListOptions{PerPage: 20}}
		workflowRuns, _, _ := client.Actions.ListWorkflowRunsByFileName(ctx, owner, repoName, workflow, opts)
		// todo use cache
		for _, run := range workflowRuns.WorkflowRuns {
			var status string
			if run.Conclusion != nil {
				status = *run.Conclusion
				if status == "success" {
					ghaRunIcon = successIcon
				} else{
					ghaRunIcon = failIcon
				}

			} else {
				status = "running"
				ghaRunIcon = runningIcon
			}

			createdAt := *run.CreatedAt
			diffMins := int(time.Since(createdAt.Time).Minutes())
			var diffString string = ""
			var diffHour int = diffMins / 60
			if diffHour != 0 {
				diffString += strconv.Itoa(diffHour) + "h"
			}
			diffString += strconv.Itoa(diffMins % 60) + "m"
			subtitle := status + " #" + strconv.Itoa(*run.RunNumber) + " " + diffString + " ago"

			wf.NewItem(*run.Name).Arg(*run.HTMLURL).Subtitle(subtitle).UID(strconv.Itoa(int(*run.ID))).Icon(ghaRunIcon).Valid(true)
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


