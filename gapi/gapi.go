package gapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Pauloo27/gmail-notifier/utils"
	"golang.org/x/oauth2"
)

func GetClient(config *oauth2.Config, tokFile string, askLogin bool) *http.Client {
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		if askLogin {
			Login(config, tokFile)
		} else {
			utils.HandleFatal("run `gmail-notifier login`. Error:", err)
		}
	}
	return config.Client(context.Background(), tok)
}

func Login(config *oauth2.Config, tokFile string) {
	tok := getTokenFromWeb(config)
	saveToken(tokFile, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		utils.HandleFatal("Unable to read authorization code", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		utils.HandleFatal("Unable to retrieve token from web", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		utils.HandleFatal("Unable to cache oauth token", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
