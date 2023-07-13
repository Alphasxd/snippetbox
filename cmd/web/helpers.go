package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
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

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	// 从 templateCache 字典中获取指定名称的模板，如果不存在，则调用 serverError() helper
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	err := ts.Execute(w, td)
	if err != nil {
		app.serverError(w, err)
	}
}