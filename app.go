package main

import (
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type App struct {
	Router      *mux.Router
	MiddleWares *Middleware
	Config      *Env
}

// shortenReq shorten require
type shortenReq struct {
	URL                 string `json:"url" validate:"required"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

// shortLinkResp short link response
type shortLinkResp struct {
	ShortLink string `json:"shortlink"`
}

// Initialize App Init
func (app *App) Initialize(env *Env) {
	logger, err := log.LoggerFromConfigAsFile("./config/seelog.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)
	defer log.Flush()
	app.Router = mux.NewRouter()
	app.MiddleWares = &Middleware{}
	app.Config = env
	app.initializeRoutes()
}

// initializeRoutes App Init Routes
func (app *App) initializeRoutes() {
	m := alice.New(app.MiddleWares.LoggingHandler, app.MiddleWares.RecoverHandler)
	app.Router.Handle("/api/shorten", m.ThenFunc(app.createShortLink)).Methods("POST")
	app.Router.Handle("/api/info", m.ThenFunc(app.getShortLinkInfo)).Methods("GET")
	app.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}", m.ThenFunc(app.redirect)).Methods("GET")
}

// createShortLink generate a short link
func (app *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		reponseWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("json r.Body failed %v", r.Body)})
		return
	}

	defer r.Body.Close()

	if err := validator.New().Struct(&req); err != nil {
		reponseWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("validate param failed %v", req)})
		return
	}

	shortLink, err := app.Config.storage.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		reponseWithError(w, StatusError{http.StatusInternalServerError, fmt.Errorf("shorten failed %v", req)})
		return
	}

	reponseWithJson(w, http.StatusOK, shortLink)
}

// getShortLinkInfo get short link information
func (app *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	shortLink := vals.Get("shortlink")

	info, err := app.Config.storage.ShortlinkInfo(shortLink)
	if err != nil {
		reponseWithError(w, err)
		return
	}

	reponseWithJson(w, http.StatusOK, info)
}

// redirect temp redirect
func (app *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	url, err := app.Config.storage.UnShorten(vars["shortlink"])
	if err != nil {
		reponseWithError(w, err)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Run app run interface
func (app *App) Run(addr string) {
	if err := http.ListenAndServe(addr, app.Router); err != nil {
		panic("Listen faild.")
	}
}

// reponseWithError analysis the error and log it
func reponseWithError(w http.ResponseWriter, statusError error) {
	switch statusError.(type) {
	case Error:
		_ = log.Errorf("http-%d-%s\n", statusError.(Error).Status(), statusError)
		reponseWithJson(w, statusError.(Error).Status(), statusError.(Error).Error())
	default:
		reponseWithJson(w, statusError.(Error).Status(), http.StatusText(statusError.(Error).Status()))
	}
}

// reponseWithJson response client with Json
func reponseWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jStr, _ := json.Marshal(payload)
	w.WriteHeader(code)
	w.Write(jStr)
}
