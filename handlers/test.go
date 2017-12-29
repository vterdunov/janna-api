package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func createTest(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")

	fmt.Fprintf(w, "POST: %s\n", login)
}

func readTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	testkey := r.FormValue("testkey")
	fmt.Fprintf(w, "FORM: %s\n", testkey)
	fmt.Fprintf(w, "GET: %s\n", name)
}

func updateTest(w http.ResponseWriter, r *http.Request) {

}

func deleteTest(w http.ResponseWriter, r *http.Request) {

}
