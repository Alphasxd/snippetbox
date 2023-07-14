package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)


func (app *application) routes() http.Handler {

	// 创建一个包含标准中间件的的中间件链，将会应用到每一个请求上。
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// 创建一个包含动态中间件的中间件链，应用到动态的路由请求上。
	dynamicMiddleware := alice.New(app.session.Enable)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/snippet/create", dynamicMiddleware.ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))


	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}