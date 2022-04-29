package main

import (
	aw "github.com/deanishe/awgo"
)


var (
	repoIcon  = &aw.Icon{Value: "icons/github-repo.png"}
)

func runFetchRepo(){
	logger.Println("Start fetch respos alfred workflow")
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

	if len(query) > 1 {
		logger.Println("query: ", query)
		wf.Filter(query)
	}
	wf.SendFeedback()
}


