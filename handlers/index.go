package handlers

import (
	"fmt"
	"net/http"
)

func index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "Hey!")
}
