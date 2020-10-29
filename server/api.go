package server

import (
	"net/http"

	"github.com/fiatjaf/ilno/ilno"
	"github.com/fiatjaf/ilno/lnurl"
	"github.com/gorilla/mux"
)

func wip(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("work in process\n"))
}

func registerRoute(router *mux.Router, ilno *ilno.ILNO) {
	router.HandleFunc("/lnurlauth", lnurl.Auth).Methods("GET").Name("lnurlauth")
	router.HandleFunc("/lnurlauth/stream", lnurl.AuthStream).Methods("GET").Name("lnurlauthstream")
	router.HandleFunc("/new", ilno.CreateComment()).Queries("uri", "{uri}").Methods("POST").Name("new")

	// single comment
	router.HandleFunc("/id/{id:[0-9]+}", ilno.ViewComment()).Methods("GET").Name("view")
	router.HandleFunc("/id/{id:[0-9]+}", ilno.EditComment()).Methods("PUT").Name("edit")
	router.HandleFunc("/id/{id:[0-9]+}/delete", ilno.DeleteComment()).Methods("POST").Name("delete")
	router.HandleFunc("/id/{id:[0-9]+}/{vote:(?:like|dislike)}", ilno.VoteComment()).Methods("POST").Name("vote")

	router.HandleFunc("/id/{id:[0-9]+}/{action:(?:edit|activate|delete)}/{key}", wip).
		Methods("GET").Name("moderate_get")
	router.HandleFunc("/id/{id:[0-9]+}/{action:(?:edit|activate|delete)}>/{key}", wip).
		Methods("POST").Name("moderate_post")

	// ping
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong\n"))
	}).Name("ping")

	// total staff
	router.HandleFunc("/latest", wip).Methods("GET").Name("latest")
	router.HandleFunc("/count", wip).Methods("GET").Name("count")
	router.HandleFunc("/count", ilno.CountComment()).Methods("POST").Name("counts")

	router.PathPrefix("/js").Handler(http.FileServer(AssetFile()))

	router.HandleFunc("/", ilno.FetchComments()).Queries("uri", "{uri}").Methods("GET").Name("fetch")
}
