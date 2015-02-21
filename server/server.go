package server

import (
	"fmt"
	"net/http"
	"github.com/zenazn/goji"
	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
	"gopkg.in/boj/redistore.v1"
	"github.com/gorilla/sessions"
	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/shiwork/favpost/model"
	"github.com/shiwork/favpost/config"
	"database/sql"
	"strconv"
)

var conf config.FavPConfig
var db *sql.DB

func Run(conf config.FavPConfig, dbInstance *sql.DB) {
	conf = conf
	db = dbInstance
	pongo2.DefaultSet.SetBaseDirectory("server/view")

	goji.Get("/", Top)
	goji.Get("/setting", Setting)
	goji.Get("/login", Login)
	goji.Get("/login/callback", LoginCallback)
//	goji.Get("/login", Login)
//	goji.Get("/login/callback", Callback)

	goji.Serve()
}

var session *sessions.Session

func InitSession(r *http.Request) {
	store, err := redistore.NewRediStore(10, "tcp", ":6379", "", []byte("secret-key"))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	store.Options = &sessions.Options{
		Domain:"127.0.0.1",
		MaxAge: 10*24*3600,
	}

	session, err = store.Get(r, "session-key")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func Top(c web.C, w http.ResponseWriter, r *http.Request) {
	InitSession(r)

	_, loginStatus := session.Values["user_id"]

	if loginStatus {
		// login済みの場合は設定画面に遷移
		http.Redirect(w, r, "/setting", 303)
	} else {
		// ログインしてない場合はログイン画面を表示
		tpl, err := pongo2.DefaultSet.FromFile("top.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tpl.ExecuteWriter(pongo2.Context{}, w)
	}
}

func Setting(c web.C, w http.ResponseWriter, r *http.Request){
	InitSession(r)

	_, loginStatus := session.Values["user_id"]
	if !loginStatus {
		// ログインしてないのでTopに飛ばす
		http.Redirect(w, r, "/", 303)
	}

	tpl, err := pongo2.DefaultSet.FromFile("setting.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteWriter(pongo2.Context{}, w)
}

// Twitterでしか認証しないので問答無用で遷移させる
func Login(c web.C, w http.ResponseWriter, r *http.Request) {
	InitSession(r)

	callbackURL := "http://127.0.0.1:8000/login/callback"
	authURL, tempCred, err := anaconda.AuthorizationURL(callbackURL)
	if err != nil {
		fmt.Println("Error: %v", err)
	}

	session.Values["temp_twitter_token"] = tempCred.Token
	session.Values["temp_twitter_secret"] = tempCred.Secret

	fmt.Println(session.Values["temp_twitter_token"])
	if err = sessions.Save(r, w); err != nil {
		fmt.Println("Error saving session: %v", err)
	}

	fmt.Println(authURL)
	http.Redirect(w, r, authURL, 303)
}

func LoginCallback(c web.C, w http.ResponseWriter, r *http.Request) {
	InitSession(r)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error! %s\n", err), http.StatusInternalServerError)
		return
	}

	fmt.Println(session.Values["temp_twitter_token"])
	verifier := r.Form["oauth_verifier"][0]

	tempCred := &oauth.Credentials{
		Token: session.Values["temp_twitter_token"].(string),
		Secret: session.Values["temp_twitter_secret"].(string),
	}
	fmt.Println(tempCred)
	fmt.Println(verifier)

	credentials, values, err := anaconda.GetCredentials(tempCred, verifier)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error! %s\n", err), http.StatusInternalServerError)
		return
	}

	if credentials == nil {
		http.Error(w, "Credentials nil", http.StatusInternalServerError)
		return
	}

	fmt.Println(values)

	user_id, _ := strconv.ParseInt(values["user_id"][0], 10, 64)


	token := model.AccessToken{
		Id: user_id,
		Token: values["oauth_token"][0],
		Secret: values["oauth_token_secret"][0],
	}
	user := model.User{
		Id: user_id,
		ScreenName: values["screen_name"][0],
		AccessToken: token,
	}

	// save user and token
	repo := model.GetUserRepository(db)
	if !repo.Exists(user.Id) {
		repo.Add(user)
	}
	repo.SaveToken(user)

	// login session
	session.Values["user_id"] = user_id
	if err = sessions.Save(r, w); err != nil {
		fmt.Println("Error saving session: %v", err)
	}

//	tpl, err := pongo2.DefaultSet.FromFile("login_callback.html")
//	tpl.ExecuteWriter(pongo2.Context{"token": *credentials}, w)

	// save user data and create login session
	// success login
	// redirect to setting
	http.Redirect(w, r, "/setting", 301)
}