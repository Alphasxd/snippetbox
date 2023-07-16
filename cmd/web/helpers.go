package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// serverError() helper 向 errorLog 写入错误信息，并向用户返回 500 Internal Server Error
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// 使用 app.errorLog.Output() 方法将错误信息写入日志，第一个参数为调用栈的深度，第二个参数为错误信息
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError() helper 发送指定的状态码和描述信息到用户端
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound() helper 使用 clientError() 方法发送 404 状态码到用户端
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash")
	td.IsAuthenticated = app.isAuthenticated(r)

	// 将身份验证信息添加到 templateData 结构中
	return td
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	// 从 templateCache 获取指定名称的模板，如果不存在，则调用 serverError() helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	// 创建一个缓冲区，然后执行模板，将渲染的结果写入缓冲区
	buf := new(bytes.Buffer)

	// 将存储在 templateData 中的动态数据写入缓冲区，同时将当前的年份信息写入缓冲区
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// 将缓冲区的内容写入到 http.ResponseWriter
	buf.WriteTo(w)
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.session.Exists(r, "authenticatedUserID")
}
