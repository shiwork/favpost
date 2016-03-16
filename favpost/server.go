package favpost

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/flosch/pongo2"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/sessions"
	"github.com/shiwork/favpost/config"
	"github.com/shiwork/favpost/model"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"gopkg.in/boj/redistore.v1"
)

var conf config.FavPConfig
var db *sql.DB

func Run(conf config.FavPConfig, dbInstance *sql.DB) {
	conf = conf
	db = dbInstance
	pongo2.DefaultSet.SetBaseDirectory(conf.TemplatePath)

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
		Domain: r.Host,
		MaxAge: 10 * 24 * 3600,
	}

	session, err = store.Get(r, "session-key")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func Top(c web.C, w http.ResponseWriter, r *http.Request) {
	InitSession(r)

	user_id, loginStatus := session.Values["user_id"]
	tpl, err := pongo2.DefaultSet.FromFile("top.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo := model.GetTweetRepository(db)
	tweets, err := repo.Find(20)
	if err != nil {
		fmt.Println(err)
	}

	if loginStatus {
		repo := model.GetUserRepository(db)
		user, _ := repo.Get(user_id.(int64))

		// login済みの場合は設定画面に遷移
		tpl.ExecuteWriter(pongo2.Context{"login": loginStatus, "user": user, "tweets": tweets}, w)
	} else {
		// ログインしてない場合はログイン画面を表示
		tpl.ExecuteWriter(pongo2.Context{"login": loginStatus, "tweets": tweets}, w)
	}
}

func Setting(c web.C, w http.ResponseWriter, r *http.Request) {
	InitSession(r)

	user_id, loginStatus := session.Values["user_id"]
	if !loginStatus {
		// ログインしてないのでTopに飛ばす
		http.Redirect(w, r, "/", 303)
	}

	tpl, err := pongo2.DefaultSet.FromFile("setting.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo := model.GetUserRepository(db)
	user, _ := repo.Get(user_id.(int64))

	tpl.ExecuteWriter(pongo2.Context{"user": user}, w)
}

// Twitterでしか認証しないので問答無用で遷移させる
func Login(c web.C, w http.ResponseWriter, r *http.Request) {
	InitSession(r)

	callbackURL := "http://" + r.Host + "/login/callback"
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
		Token:  session.Values["temp_twitter_token"].(string),
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
		Id:     user_id,
		Token:  values["oauth_token"][0],
		Secret: values["oauth_token_secret"][0],
	}
	user := model.User{
		Id:          user_id,
		ScreenName:  values["screen_name"][0],
		AccessToken: token,
	}

	// save user and token
	repo := model.GetUserRepository(db)
	err = repo.Add(user)
	if err != nil {
		fmt.Println("Error: %v", err)
	}
	err = repo.SaveToken(user)
	if err != nil {
		fmt.Println("Error: %v", err)
	}

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
