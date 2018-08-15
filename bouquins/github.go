package bouquins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GithubProvider implements OAuth2 client with github.com
type GithubProvider string

type githubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

func init() {
	Providers = append(Providers, GithubProvider("github"))
}

// Name returns name of provider
func (p GithubProvider) Name() string {
	return string(p)
}

// Label returns label of provider
func (p GithubProvider) Label() string {
	return "Github"
}

// Icon returns icon CSS class for provider
func (p GithubProvider) Icon() string {
	return "githubicon"
}

// Config returns OAuth configuration for this provider
func (p GithubProvider) Config(conf *Conf) *oauth2.Config {
	for _, c := range conf.ProvidersConf {
		if c.Name == p.Name() {
			return &oauth2.Config{
				ClientID:     c.ClientID,
				ClientSecret: c.ClientSecret,
				Scopes:       []string{"user:email"},
				Endpoint:     github.Endpoint,
			}
		}
	}
	return nil
}

// GetUser returns github primary email
func (p GithubProvider) GetUser(token *oauth2.Token) (string, error) {
	apiReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	apiReq.Header.Add("Accept", "application/vnd.github.v3+json")
	apiReq.Header.Add("Authorization", "token "+token.AccessToken)
	client := &http.Client{}
	response, err := client.Do(apiReq)
	defer response.Body.Close()
	if err != nil {
		log.Println("Auth error", err)
		return "", fmt.Errorf("Authentification error")
	}

	dec := json.NewDecoder(response.Body)
	var emails []githubEmail
	err = dec.Decode(&emails)
	if err != nil {
		log.Println("Error reading github API response", err)
		return "", fmt.Errorf("Error reading github API response")
	}
	var userEmail string
	for _, email := range emails {
		if email.Primary && email.Verified {
			userEmail = email.Email
		}
	}
	log.Println("User email:", userEmail)
	return userEmail, nil
}

func findProvider(name string) OAuth2Provider {
	for _, p := range Providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
