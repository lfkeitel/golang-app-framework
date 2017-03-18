package controllers

import (
	"net/http"

	"github.com/lfkeitel/golang-app-framework/src/utils"
)

type HelloWorld struct {
	e *utils.Environment
}

func NewHelloWorldController(e *utils.Environment) *HelloWorld {
	return &HelloWorld{e: e}
}

// Dev mode route handlers
func (d *HelloWorld) Main(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}
