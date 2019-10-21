package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)
import log "github.com/cihub/seelog"

type App struct {
	Router *mux.Router
}

type shortenReq struct {
	URL string `json:"url" validate:"nonzero"`
	ExpirationInMinutes int64 `json:"expiration_in_minutes" validate:"min=0"`
}

type shortLinkResp struct {
	ShortLink string `json:"shortlink"`
}

// App Init
func (app *App) Initialize() {
	logger, err := log.LoggerFromConfigAsFile("./config/seelog.xml")
	if err != nil {
		log.Critical("err parsing config log file", err)
		return
	}
	log.ReplaceLogger(logger)
	defer log.Flush()
	app.Router = mux.NewRouter()
	app.initializeRoutes()
}

func(app *App) initializeRoutes() {
	app.Router.HandleFunc("/api/shorten", app.createShortLink).Methods("POST")
	app.Router.HandleFunc("/api/info", app.getShortLinkInfo).Methods("GET")
	app.Router.HandleFunc("/{shorten:[a-zA-Z0-9]{1,11}}", app.redirect).Methods("GET")
}

func(app *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Json decoder failed.")
		return
	}

	defer r.Body.Close()

	if err := validator.New().Struct(req); err != nil {
		return
	}

	fmt.Println(req)
}

func(app *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	s := vals.Get("shortlink")

	fmt.Println(s)
}

func(app *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("%s\n", vars["shortlink"])
}

func (app *App) Run(addr string) {
	if err := http.ListenAndServe(addr, app.Router); err != nil {
		panic("Listen faild.")
	}
}