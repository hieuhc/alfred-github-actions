package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type RepoWorkflowItem struct {
	Owner string
	Name string
	Description string
	UID string
	HTMLURL string
 
}

func fetchRepo(context context.Context, token string) ([]RepoWorkflowItem, error){
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context, ts)
	client := github.NewClient(tc)


	pageNum := 0
	repoItems := make([]RepoWorkflowItem, 0)
	for {
		opts := &github.RepositoryListOptions{ListOptions: github.ListOptions{Page: pageNum, PerPage: 100}}
		repos, _, err := client.Repositories.List(context, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		if len(repos) == 0{
			break
		}
		// todo use cache
		for _, p := range repos {
			// todo shorten
			var desc string
			if p.Description != nil {
				desc = *p.Description
			} else {
				desc = ""
			}
			repoItems = append(repoItems, RepoWorkflowItem{*p.Owner.Login, *p.Name, desc, strconv.Itoa(int(*p.ID)), *p.HTMLURL})
		}
		pageNum += 1
	}
	return repoItems, nil
}

func runEntry(){
	ctx := context.Background()

	// try login with this token
	if ghToken != "" {
		fmt.Println("Attemping logging with provided token")
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: ghToken},
		)
		oauthClient := oauth2.NewClient(ctx, tokenSource)
		ghClient := github.NewClient(oauthClient)
		_, _, err := ghClient.Users.Get(ctx, "")
		

		if err != nil{
			wf.FatalError(err)
		} else {
			// save token to keychain
			err = saveToken(ghToken)
			if err !=nil {
				wf.FatalError(err)
			}
			logger.Printf("Login succeedded")
		}
	} else {
		logger.Printf("Refreshing the list of repositories")
		var err error
		ghToken, err = getToken()
		if err != nil {
			wf.Fatal(err.Error())
		}

	}
	repos, err := fetchRepo(ctx, ghToken)
	if err != nil {
		wf.Fatal(err.Error())
	}
	err = wf.Cache.StoreJSON(cachedGithubRepos, repos)
	if err != nil {
		wf.Fatal(err.Error())
	}
	logger.Printf("List of repos are cached")
}


