package main

import (
	"log"
	"os"
	aw "github.com/deanishe/awgo"
)

// TODO reuse from login.go
const (
	cachedGithubRepos = "cached-github-repositories.json"
)

var (
	logger = log.New(os.Stderr, "prefixLogger", log.LstdFlags)
	repoIcon  = &aw.Icon{Value: "icons/github-repo.png"}
	wf *aw.Workflow
)

// TODO reuse from login.go
type RepoWorkflowItem struct {
	Owner string
	Name string
	Description string
	UID string
	HTMLURL string
}

func init(){
	wf = aw.New()
}

func run(){
	logger.Println("Start fetch respos alfred workflow")
	args := wf.Args()

	var repos []RepoWorkflowItem
	if !wf.Cache.Exists(cachedGithubRepos) {
		wf.Fatal("No repos cached, please run -refresh")
	}
	if err := wf.Cache.LoadJSON(cachedGithubRepos, &repos); err != nil {
		wf.Fatal(err.Error())
	}


	for _, repoItem := range repos {
		repoFullName := repoItem.Owner + "/" + repoItem.Name
		wf.NewItem(repoFullName).Arg(repoFullName).Subtitle(repoItem.Description).UID(repoItem.UID).Icon(repoIcon).Valid(true).NewModifier("cmd").Arg(repoItem.HTMLURL)
	}

	if len(args) > 0 {
		logger.Println("query: ", args[0])
		wf.Filter(args[0])
	}
	wf.SendFeedback()
}


func main(){
	wf.Run(run)
}


