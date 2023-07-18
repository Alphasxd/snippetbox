package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Alphasxd/snippetbox/pkg/models"
	
	"github.com/justinas/nosurf"
)

func secureHeaders(next http.Handler) http.Handler {

	// 设置希望添加到响应中的标准的安全标头
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {

	// logRequest 中间件将所有请求的远程地址和 HTTP 方法记录到应用的日志中
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {

	// 恢复 panic，如果发生 panic，则将堆栈跟踪信息写入日志，可以防止应用程序崩溃
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {

	// 鉴定用户是否已经通过身份验证
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果用户没有通过身份验证，将用户重定向到登录页面
		if !app.isAuthenticated(r) {
			app.session.Put(r, "redirectPathAfterLogin", r.URL.Path)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		
		// 否则，将 header 中的 Cache-Control 字段设置为 no-store，这样用户每次访问受保护的页面时，都会向服务器发送请求
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查用户是否在存在 session 中，如果不存在则调用 next.ServeHTTP() 方法
		exists := app.session.Exists(r, "authenticatedUserID")
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		// 从 session 中获取用户的 ID，然后从数据库中检索相关的用户记录 
		// 如果没有找到匹配的记录，或者用户处于非活动状态，则将 session 从用户的浏览器中删除并调用 next.ServeHTTP() 方法
		user, err := app.users.Get(app.session.GetInt(r, "authenticatedUserID"))
		if errors.Is(err, models.ErrNoRecord) || !user.Active {
			app.session.Remove(r, "authenticatedUserID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}

		// 
		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	}) 
}