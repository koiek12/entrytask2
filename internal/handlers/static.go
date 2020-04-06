package handlers

import (
	"net/http"
	"text/template"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("../../web/template/login.html")
	t.Execute(w, nil)
}
