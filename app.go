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
	Middlewares *Middleware
}

// shorten require
type shortenReq struct {
	URL                 string `json:"url" validate:"required"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

// short link response
type shortLinkResp struct {
	ShortLink string `json:"shortlink"`
}

// App Init
func (app *App) Initialize() {
	logger, err := log.LoggerFromConfigAsFile("./config/seelog.xml")
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)
	defer log.Flush()
	app.Router = mux.NewRouter()
	app.Middlewares = &Middleware{}
	app.initializeRoutes()
}

// App Init Routes
func (app *App) initializeRoutes() {
	m := alice.New(app.Middlewares.LoggingHandler, app.Middlewares.RecoverHandler)
	app.Router.Handle("/api/shorten", m.ThenFunc(app.createShortLink)).Methods("POST")
	app.Router.Handle("/api/info", m.ThenFunc(app.getShortLinkInfo)).Methods("GET")
	app.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}", m.ThenFunc(app.redirect)).Methods("GET")
}

// generate a short link
func (app *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	fmt.Println(r.Body)
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

	fmt.Println(req)
}

// get short link information
func (app *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	s := vals.Get("shortlink")

	fmt.Println(s)
}

// temp redirect
func (app *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("%s\n", vars["shortlink"])
}

func (app *App) Run(addr string) {
	if err := http.ListenAndServe(addr, app.Router); err != nil {
		panic("Listen faild.")
	}
}

func reponseWithError(w http.ResponseWriter, statusError error) {
	switch statusError.(type) {
	case Error:
		_ = log.Errorf("http-%d-%s\n", statusError.(Error).Status(), statusError)
		reponseWithJson(w, statusError.(Error).Status(), statusError.(Error).Error())
	default:
		reponseWithJson(w, statusError.(Error).Status(), http.StatusText(statusError.(Error).Status()))
	}
}

func reponseWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	jStr, _ := json.Marshal(payload)
	w.WriteHeader(code)
	w.Write(jStr)
}
