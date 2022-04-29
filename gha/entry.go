package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v41/github"
	"github.com/keybase/go-keychain"
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
			item := keychain.NewItem()
			item.SetSecClass(keychain.SecClassGenericPassword)
			item.SetService(keychainService)
			item.SetData([]byte(ghToken))
			keychainErr := keychain.AddItem(item)

			if keychainErr == keychain.ErrorDuplicateItem {
				logger.Printf("Updating Github PAT in keychain")
				errDelete := keychain.DeleteItem(item)
				errAdd := keychain.AddItem(item)
				if (errDelete != nil || errAdd != nil){
					wf.Fatal("Failed updating the PAT")
				}
			} else if (keychainErr != nil){
				wf.Fatal("Failed setting the PAT")
			}
			logger.Printf("Login succeedded")
		}
	} else {
		logger.Printf("Refreshing the list of repositories")
		tokenHolder := keychain.NewItem()
		tokenHolder.SetSecClass(keychain.SecClassGenericPassword)
		tokenHolder.SetService(keychainService)
		tokenHolder.SetMatchLimit(keychain.MatchLimitOne)
		tokenHolder.SetReturnData(true)
		results, err := keychain.QueryItem(tokenHolder)

		if err != nil {
			wf.Fatal(err.Error())
		} else if len(results) != 1 {
			wf.Fatal("Github PAT not found in keychain")
		} else {
			ghToken = string(results[0].Data)
			logger.Println("Found Github PAT in keychain")
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


