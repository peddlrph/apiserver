package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
	conf   *Config
}

type Config struct {
	Port      string `json:"port"`
	Host      string `json:"host"`
	Token     string `json:"token"`
	DBConnect string `json:"db_connect"`
}

//func (a *App) Initialize(user, password, dbname string) {
func (a *App) Initialize() {

	//var config Config

	file, err := os.Open("./config.json")
	if err != nil {
		fmt.Println(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&a.conf)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(a.conf)

	connectionString := a.conf.DBConnect

	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/message/last", a.getLastMessage).Methods("GET")
	a.Router.HandleFunc("/messages", a.getMessages).Methods("GET")
	a.Router.HandleFunc("/message", a.createMessage).Methods("POST")
	a.Router.HandleFunc("/message/{id:[0-9]+}", a.getMessage).Methods("GET")
	//a.Router.Use(loggingMiddleware)
	//a.Router.Use(a.AuthHandler)
	a.Router.Use(a.LogHandler)
}

func (a *App) Run(addr string) {
	fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), "Running HTTP on port "+a.conf.Port)
	log.Fatal(http.ListenAndServe(":"+a.conf.Port, a.Router))
}

func (a *App) LogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now().Format("2006-01-02 03:04:05 PM"), r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func (a *App) AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer "+a.conf.Token {
			next.ServeHTTP(w, r)
		} else {
			fmt.Println("Not Authorized")
		}
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (a *App) getLastMessage(w http.ResponseWriter, r *http.Request) {

	m := message{}
	if err := m.getLastMessage(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Message not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, m)
}

func (a *App) getMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	m := message{Id: id}
	if err := m.getMessage(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Message not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, m)
}

func (a *App) getMessages(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	messages, err := getMessages(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, messages)
}

func (a *App) createMessage(w http.ResponseWriter, r *http.Request) {
	var m message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := m.createMessage(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, m)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
