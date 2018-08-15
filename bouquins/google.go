package bouquins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleProvider implements OAuth2 client with google account
type GoogleProvider string

type googleTokenInfo struct {
	IssuedTo      string `json:"issued_to"`
	Audience      string `json:"audience"`
	UserID        string `json:"user_id"`
	Scope         string `json:"scope"`
	ExpiresIn     int64  `json:"expires_in"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	AccessType    string `json:"access_type"`
}

func init() {
	Providers = append(Providers, GoogleProvider("google"))
}

// Name returns name of provider
func (p GoogleProvider) Name() string {
	return string(p)
}

// Label returns label of provider
func (p GoogleProvider) Label() string {
	return "Google"
}

// Icon returns icon path for provider
func (p GoogleProvider) Icon() string {
	return "googleicon"
}

// Config returns OAuth configuration for this provider
func (p GoogleProvider) Config(conf *Conf) *oauth2.Config {
	for _, c := range conf.ProvidersConf {
		if c.Name == p.Name() {
			return &oauth2.Config{
				ClientID:     c.ClientID,
				ClientSecret: c.ClientSecret,
				Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
				Endpoint:     google.Endpoint,
				RedirectURL:  conf.ExternalURL + URLCallback,
			}
		}
	}
	return nil
}

// GetUser returns github primary email
func (p GoogleProvider) GetUser(token *oauth2.Token) (string, error) {
	apiRes, err := http.Post("https://www.googleapis.com/oauth2/v2/tokeninfo?access_token="+token.AccessToken, "application/json", nil)
	defer apiRes.Body.Close()
	if err != nil {
		log.Println("Auth error", err)
		return "", fmt.Errorf("Authentification error")
	}
	dec := json.NewDecoder(apiRes.Body)
	var tokenInfo googleTokenInfo
	err = dec.Decode(&tokenInfo)
	if err != nil {
		log.Println("Error reading google API response", err)
		return "", fmt.Errorf("Error reading google API response")
	}
	var userEmail string
	if tokenInfo.VerifiedEmail {
		userEmail = tokenInfo.Email
	}
	log.Println("User email:", userEmail)
	return userEmail, nil
}
