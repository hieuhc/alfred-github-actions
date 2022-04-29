package main

import (
	"github.com/keybase/go-keychain"
)

func saveToken(token string) (error){
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(keychainService)
	item.SetData([]byte(token))
	keychainErr := keychain.AddItem(item)

	if keychainErr == keychain.ErrorDuplicateItem {
		logger.Printf("Updating Github PAT in keychain")
		errDelete := keychain.DeleteItem(item)
		if errDelete != nil {
			return errDelete
		}
		errAdd := keychain.AddItem(item)
		if errAdd != nil{
			return errAdd
		}
	} else if (keychainErr != nil){
		wf.Fatal("Failed setting the PAT")
	}
	return nil
}

func getToken() (string, error){
	logger.Printf("Refreshing the list of repositories")
	tokenHolder := keychain.NewItem()
	tokenHolder.SetSecClass(keychain.SecClassGenericPassword)
	tokenHolder.SetService(keychainService)
	tokenHolder.SetMatchLimit(keychain.MatchLimitOne)
	tokenHolder.SetReturnData(true)
	results, err := keychain.QueryItem(tokenHolder)

	var token string
	if err != nil || len(results) != 1{
		return "", err
	} else {
		token = string(results[0].Data)
		logger.Println("Found Github PAT in keychain")
	}
	return token, nil
}
