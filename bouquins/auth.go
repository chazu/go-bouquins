package bouquins

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

const (
	alphanums            = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	sessionName          = "bouquins"
	sessionOAuthState    = "oauthState"
	sessionOAuthProvider = "provider"
	sessionUser          = "username"

	pProvider = "provider"
)

var (
	// Providers contains OAuth2 providers implementations
	Providers []OAuth2Provider
)

// LoginModel is login page model
type LoginModel struct {
	Model
	Providers []OAuth2Provider
}

// NewLoginModel constructor for LoginModel
func (app *Bouquins) NewLoginModel(req *http.Request) *LoginModel {
	return &LoginModel{*app.NewModel("Authentification", "provider", req), Providers}
}

// OAuth2Provider allows to get a user from an OAuth2 token
type OAuth2Provider interface {
	GetUser(token *oauth2.Token) (string, error)
	Config(conf *Conf) *oauth2.Config
	Name() string
	Label() string
	Icon() string
}

// generates a 16 characters long random string
func securedRandString() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = alphanums[rand.Intn(len(alphanums))]
	}
	return string(b)
}

// Session returns current session
func (app *Bouquins) Session(req *http.Request) *sessions.Session {
	session, _ := app.Cookies.Get(req, sessionName)
	return session
}

// Username returns logged in username
func (app *Bouquins) Username(req *http.Request) string {
	username := app.Session(req).Values[sessionUser]
	if username != nil {
		return username.(string)
	}
	return ""
}

// SessionSet sets a value in session
func (app *Bouquins) SessionSet(name string, value string, res http.ResponseWriter, req *http.Request) {
	session := app.Session(req)
	session.Values[name] = value
	session.Save(req, res)
}

// LoginPage redirects to OAuth login page (github)
func (app *Bouquins) LoginPage(res http.ResponseWriter, req *http.Request) error {
	provider := req.URL.Query().Get(pProvider)
	oauth := app.OAuthConf[provider]
	if oauth != nil {
		app.SessionSet(sessionOAuthProvider, provider, res, req)
		state := securedRandString()
		app.SessionSet(sessionOAuthState, state, res, req)
		url := oauth.AuthCodeURL(state)
		log.Println("OAuth redirect", url)
		http.Redirect(res, req, url, http.StatusTemporaryRedirect)
		return nil
	}
	// choose provider
	return app.render(res, tplProvider, app.NewLoginModel(req))
}

// LogoutPage logout connected user
func (app *Bouquins) LogoutPage(res http.ResponseWriter, req *http.Request) error {
	app.SessionSet(sessionUser, "", res, req)
	return RedirectHome(res, req)
}

// CallbackPage handle OAuth 2 callback
func (app *Bouquins) CallbackPage(res http.ResponseWriter, req *http.Request) error {
	savedState := app.Session(req).Values[sessionOAuthState]
	providerParam := app.Session(req).Values[sessionOAuthProvider]
	if savedState == "" || providerParam == "" {
		return fmt.Errorf("missing oauth data")
	}
	providerName := providerParam.(string)
	oauth := app.OAuthConf[providerName]
	provider := findProvider(providerName)
	if oauth == nil || provider == nil {
		return fmt.Errorf("missing oauth configuration")
	}
	app.SessionSet(sessionOAuthState, "", res, req)
	app.SessionSet(sessionOAuthProvider, "", res, req)
	state := req.FormValue("state")
	if state != savedState {
		return fmt.Errorf("invalid oauth state, expected '%s', got '%s'", "state", state)
	}
	code := req.FormValue("code")
	token, err := oauth.Exchange(oauth2.NoContext, code)
	if err != nil {
		return fmt.Errorf("Code exchange failed with '%s'", err)
	}
	userEmail, err := provider.GetUser(token)
	if err != nil {
		return err
	}
	user, err := Account(userEmail)
	if err != nil {
		log.Println("Error loading user", err)
		return fmt.Errorf("Unknown user")
	}
	app.SessionSet(sessionUser, user.DisplayName, res, req)
	log.Println("User logged in", user.DisplayName)
	return RedirectHome(res, req)
}
