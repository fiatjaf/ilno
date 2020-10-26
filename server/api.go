package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"wrong.wang/x/go-isso/isso"
	"wrong.wang/x/go-isso/lnurl"
)

func workInProcess(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("work in process\n"))
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

func registerRoute(router *mux.Router, isso *isso.ISSO) {
	router.HandleFunc("/lnurlauth", lnurl.Auth).Methods("GET").Name("lnurlauth")
	router.HandleFunc("/lnurlauth/stream", lnurl.AuthStream).Methods("GET").Name("lnurlauthstream")
	router.HandleFunc("/new", isso.CreateComment()).Queries("uri", "{uri}").Methods("POST").Name("new")

	// single comment
	router.HandleFunc("/id/{id:[0-9]+}", isso.ViewComment()).Methods("GET").Name("view")
	router.HandleFunc("/id/{id:[0-9]+}", isso.EditComment()).Methods("PUT").Name("edit")
	router.HandleFunc("/id/{id:[0-9]+}", isso.DeleteComment()).Methods("DELETE").Name("delete")
	router.HandleFunc("/id/{id:[0-9]+}/{vote:(?:like|dislike)}", isso.VoteComment()).Methods("POST").Name("vote")

	router.HandleFunc("/id/{id:[0-9]+}/{action:(?:edit|activate|delete)}/{key}", workInProcess).
		Methods("GET").Name("moderate_get")
	router.HandleFunc("/id/{id:[0-9]+}/{action:(?:edit|activate|delete)}>/{key}", workInProcess).
		Methods("POST").Name("moderate_post")
	router.HandleFunc("/id/{id:[0-9]+}/unsubscribe/{email}/{key}>", workInProcess).
		Methods("GET").Name("unsubscribe")

	// functional
	router.HandleFunc("/demo", workInProcess).Methods("GET").Name("demo")

	// amdin staff
	router.HandleFunc("/admin", workInProcess).Methods("GET").Name("admin")
	router.HandleFunc("/login", workInProcess).Methods("POST").Name("login")

	// ping
	router.HandleFunc("/ping", ping).Name("ping")

	// total staff
	router.HandleFunc("/latest", workInProcess).Methods("GET").Name("latest")
	router.HandleFunc("/count", workInProcess).Methods("GET").Name("count")
	router.HandleFunc("/count", isso.CountComment()).Methods("POST").Name("counts")

	router.PathPrefix("/js").Handler(http.FileServer(AssetFile()))

	router.HandleFunc("/", isso.FetchComments()).Queries("uri", "{uri}").Methods("GET").Name("fetch")
}
