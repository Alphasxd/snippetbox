package main

import (
	"net/http"

	"github.com/Alphasxd/snippetbox/ui"
	
	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)


func (app *application) routes() http.Handler {

	// 创建一个包含标准中间件的的中间件链，将会应用到每一个请求上。
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// 创建一个包含动态中间件的中间件链，应用到动态的路由请求上。
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/about", dynamicMiddleware.ThenFunc(app.about))

	mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))

	fileServer := http.FileServer(http.FS(ui.Files))
	mux.Get("/static/", fileServer)

	return standardMiddleware.Then(mux)
}