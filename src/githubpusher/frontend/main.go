package main

import (
	"encoding/json"
	"flag"
	"github.com/google/go-github/github"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"githubpusher/config"
	"githubpusher/db"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
)

func handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := gOAuthConf.AuthCodeURL(gOAuthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGitHubLogout(w http.ResponseWriter, r *http.Request) {
	token := getAuthTokenFromRequest(r)
	if token != nil {
		d, _ := tokenToJSON(token)
		db.RemoveUserByToken(d)
		session, err := CookieStore.Get(r, "auth-token")
		if err == nil {
			session.Values["token"] = "removed"
			session.Save(r, w)
		}
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != gOAuthStateString {
		log.Println("invalid oauth state, expected '%s', got '%s'\n", gOAuthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := gOAuthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println("gOAuthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	session, err := CookieStore.Get(r, "auth-token")
	if err != nil {
		log.Println("CookieStore can't get. '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	d, _ := tokenToJSON(token)
	session.Values["token"] = d
	session.Save(r, w)

	http.Redirect(w, r, "/install", http.StatusTemporaryRedirect)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	token := getAuthTokenFromRequest(r)

	if token != nil {
		oauthClient := gOAuthConf.Client(oauth2.NoContext, token)
		client := github.NewClient(oauthClient)
		_, _, err := client.Users.Get("")
		if err == nil {
			http.Redirect(w, r, "/manage", http.StatusTemporaryRedirect)
			return
		}
	}

	config := config.ReadConfig()
	t := template.New("index.html")
	index, err := t.ParseFiles(path.Join(config.PackagePath, "templates/index.html"))
	if err != nil {
		log.Fatal("template.ParseFiles: ", err)
	}

	model := map[string]interface{}{
		"Name": "Hi!",
	}
	err = index.Execute(w, model)
	if err != nil {
		log.Println("There was an error:", err)
	}
}

func handleInstall(w http.ResponseWriter, r *http.Request) {
	config := config.ReadConfig()
	t := template.New("install.html")
	index, err := t.ParseFiles(path.Join(config.PackagePath, "templates/install.html"))
	if err != nil {
		log.Fatal("template.ParseFiles: ", err)
	}

	model := map[string]interface{}{
		"Name": "Hi!",
	}

	err = index.Execute(w, model)
	if err != nil {
		log.Println("There was an error:", err)
	}

}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	token := getAuthTokenFromRequest(r)

	oauthClient := gOAuthConf.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	user, _, err := client.Users.Get("")
	if err != nil {
		log.Println("client.Users.Get() faled with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	starred, _, err := client.Activity.ListStarred("", nil)
	if err != nil {
		log.Println("Activity.ListStarred returned error:", err)
		return
	}

	//TODO add something to go update the user's repos ever so often
	repos := []string{}
	for _, e := range starred {
		repos = append(repos, *e.Repository.HTMLURL)
	}

	d, _ := tokenToJSON(token)

	log.Println("token: ", d)
	db.AddUsersToDB(&db.User{*user.Name, d, repos, r.FormValue("endpoint"), r.FormValue("subscription"), nil})
}

func handleManage(w http.ResponseWriter, r *http.Request) {
	config := config.ReadConfig()

	t := template.New("manage.html")
	index, err := t.ParseFiles(path.Join(config.PackagePath, "templates/manage.html"))
	if err != nil {
		log.Fatal("template.ParseFiles: ", err)
	}

	token := getAuthTokenFromRequest(r)
	d, _ := tokenToJSON(token)
	u := db.GetUserByToken(d)

	model := map[string]interface{}{
		"Name":          u.Name,
		"Repos":         u.Repos,
		"Count":         len(u.PushEvents),
		"PendingEvents": u.PushEvents,
	}
	err = index.Execute(w, model)
	if err != nil {
		log.Println("There was an error:", err)
	}
}

var gOAuthConf *oauth2.Config
var gOAuthStateString string

func setupOAuth(config *config.ServerConfig) {
	gOAuthConf = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		// select level of access you want https://developer.github.com/v3/oauth/#scopes
		Scopes:   []string{"user:email", "repo"},
		Endpoint: githuboauth.Endpoint,
	}
	// random string for oauth2 API calls to protect against CSRF
	gOAuthStateString = config.OAuthState
}

func tokenToJSON(token *oauth2.Token) (string, error) {
	if d, err := json.Marshal(token); err != nil {
		return "", err
	} else {
		return string(d), nil
	}
}

func tokenFromJSON(jsonStr string) (*oauth2.Token, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(jsonStr), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func getAuthTokenFromRequest(r *http.Request) *oauth2.Token {
	session, err := CookieStore.Get(r, "auth-token")
	if err != nil {
		log.Println("CookieStore can't get. '%s'\n", err)
		return nil
	}

	tokenStr := session.Values["token"]

	if tokenStr == nil {
		log.Println("no token\n")
		return nil
	}
	token, _ := tokenFromJSON(tokenStr.(string))
	return token
}

var CookieStore *sessions.CookieStore

func setupCookie(config *config.ServerConfig) {
	CookieStore = sessions.NewCookieStore([]byte(config.CookieSecret))
	CookieStore.Options = &sessions.Options{
		MaxAge: 86400 * 365,
		Secure: true,
		Path:   "/",
	}
}

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Hello")
	flag.Parse()

	config := config.ReadConfig()
	setupOAuth(config)
	setupCookie(config)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleGitHubLogin)
	http.HandleFunc("/logout", handleGitHubLogout)

	http.HandleFunc("/github_oauth_cb", handleGitHubCallback)
	http.HandleFunc("/manage", handleManage)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/install", handleInstall)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(path.Join(config.PackagePath, "static")))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(path.Join(config.PackagePath, "static")))))
	http.Handle("/manifest.json", http.FileServer(http.Dir(path.Join(config.PackagePath, "static"))))
	http.Handle("/sw.js", http.FileServer(http.Dir(path.Join(config.PackagePath, "static"))))

	log.Println("Listening on", config.Port)
	err := http.ListenAndServeTLS(":"+config.Port,
		config.CertFilename,
		config.KeyFilename,
		context.ClearHandler(http.DefaultServeMux))

	log.Println("Exiting... ", err)
}
