package main

import "net/http"

// The routes() method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("/", app.homePage)
	mux.HandleFunc("/post/create", app.postCreate)
	// mux.HandleFunc("/snippet/create", func(w http.ResponseWriter, r *http.Request) {
	// 	app.requireAuthenticatedUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// 		app.snippetCreate(w, r)
	// 	})).ServeHTTP(w, r)
	// })
	mux.HandleFunc("/user/signup", app.userSignup)
	mux.HandleFunc("/post/view", app.postView)
	mux.HandleFunc("/user/login", app.userLogin)
	mux.HandleFunc("/user/logout", app.userLogout)
	mux.HandleFunc("/myposts", app.userPosts)
	return app.logRequest(secureHeaders(mux))
}
