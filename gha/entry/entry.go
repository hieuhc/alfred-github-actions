package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v41/github"
	"github.com/keybase/go-keychain"
	"golang.org/x/oauth2"
)

const (
	cachedGithubRepos = "cached-github-repositories.json"
)

var (
	logger = log.New(os.Stderr, "prefixLogger", log.LstdFlags)
	wf *aw.Workflow
	argToken string
)

func init(){
	wf = aw.New()
	flag.StringVar(&argToken, "login", "", "Github Access Token")
}

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

func run(){
	fmt.Println("Start alfred workflow")
	wf.Args()
	flag.Parse()
	ctx := context.Background()
	keychainService := "alfred_gha"

	// try login with this token
	if argToken != "" {
		fmt.Println("argToken: ", argToken, "!!")
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: argToken},
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
			item.SetData([]byte(argToken))
			keychainErr := keychain.AddItem(item)

			if keychainErr == keychain.ErrorDuplicateItem {
				// TODO override old key
				logger.Printf("Github PAT duplicated in keychain")
			}
			logger.Printf("Login succeedded")
		}
	} else {
		logger.Printf("Refreshing the list of repositories")
		token := keychain.NewItem()
		token.SetSecClass(keychain.SecClassGenericPassword)
		token.SetService(keychainService)
		token.SetMatchLimit(keychain.MatchLimitOne)
		token.SetReturnData(true)
		results, err := keychain.QueryItem(token)

		if err != nil {
			wf.Fatal(err.Error())
		} else if len(results) != 1 {
			wf.Fatal("Github PAT not found in keychain")
		} else {
			argToken = string(results[0].Data)
			logger.Println("Found Github PAT in keychain")
		}
	}
	repos, err := fetchRepo(ctx, argToken)
	if err != nil {
		wf.Fatal(err.Error())
	}
	err = wf.Cache.StoreJSON(cachedGithubRepos, repos)
	logger.Println("datadir", wf.DataDir())
	logger.Println("cachedir", wf.Cache.Dir)
	if err != nil {
		wf.Fatal(err.Error())
	}
	logger.Printf("List of repos are cached")
}

func main(){
	wf.Run(run)
}


