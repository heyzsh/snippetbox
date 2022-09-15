package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// It will use the custom 404 handler
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Test route
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)
	// No protection
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.showLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.doLogin))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.showSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.doSignup))

	protected := dynamic.Append(app.requireAuthentication)
	// Protected
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.doSignout))
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.showSnippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.doSnippetCreate))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
