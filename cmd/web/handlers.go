package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/Alphasxd/snippetbox/pkg/models"
)

// handler 是满足 http.Handler 接口中的 ServeHTTP() 方法的任何类型，譬如string、struct或者函数等其他类型
// 前提是需要满足 serveHTTP() 方法的签名：ServeHTTP(http.ResponseWriter, *http.Request)
// 但以 func (variable type) ServeHTTP(http.ResponseWriter, *http.Request) 的方式实现 Handler 接口过于繁琐
// 在实际开发中，我们可以使用 http.HandlerFunc 类型来简化这个过程
// 先定义一个符合 ServeHTTP() 函数签名的函数，譬如 func home(w http.ResponseWriter, r *http.Request) {}
// 然后使用如下的方式将其转换为一个 Handler 对象
// mux := http.NewServeMux()
// mux.Handle("/path", http.HandlerFunc(home))
// Go提供了语法糖，上述语法还可以进一步简化为：
// mux := http.NewServeMux()
// mux.HandleFunc("/path", home)

// 定义一个 home 处理器函数
// 修改 home 处理器函数，使其能够接收一个名为 app 的参数，该参数的类型是 application 结构体指针
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// 检查当前请求的 URL Path 是否与 "/" 匹配，如果不匹配则调用 http.NotFound() 函数
	if r.URL.Path != "/" {
		// 调用 notFound() helper
		app.notFound(w)
		return
	}

	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := &templateData{Snippets: s}

	files := []string{
		"./ui/html/home.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}

	// 使用 ParseFiles() 函数加载 home.page.tmpl 文件到一个模板集合中
	ts, err := template.ParseFiles(files...)
	if err != nil {
		// 调用 serverError() helper
		app.serverError(w, err)
		return
	}
	// 调用 Execute() 方法将模板传递给 http.ResponseWriter
	// Execute() 方法接收两个参数：一个 http.ResponseWriter 和一个 template 数据对象
	err = ts.Execute(w, data)
	if err != nil {
		// 同样调用 serverError() helper
		app.serverError(w, err)
	}
}

// 定义一个 showSnippet 处理器函数
// 修改 showSnippet 处理器函数，使其能够接收一个名为 app 的参数，该参数的类型是 application 结构体指针
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// 使用 r.URL.Query().Get() 方法获取 "id" 查询字符串参数的值
	id, err := strconv.Atoi(r.URL.Query().Get("id"))

	// 如果参数不存在或者不是一个有效的数字，则返回一个 404 Not Found 响应
	if err != nil || id < 1 {
		// 调用 notFound() helper
		app.notFound(w)
		return
	}

	s, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := &templateData{Snippet: s}

	files := []string{
		"./ui/html/show.page.tmpl",
		"./ui/html/base.layout.tmpl",
		"./ui/html/footer.partial.tmpl",
	}
	
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = ts.Execute(w, data)
	if err != nil {
		app.serverError(w, err)
	}

}

// 定义一个 createSnippet 处理器函数
// 修改 createSnippet 处理器函数，使其能够接收一个名为 app 的参数，该参数的类型是 application 结构体指针
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	
	// 使用 r.Method 变量检查请求的方法是否为 POST
	if r.Method != http.MethodPost {
		
		// 使用 Header().Set() 方法设置响应头部的 "Allow: POST"
		// 允许使用 POST 方法的请求通过，注意必须在调用 w.WriteHeader() 方法之前调用该方法
		// Cautious handlers should read the Request.Body first, and then reply.
		w.Header().Set("Allow", http.MethodPost)
		// 调用 clientError() helper，传入 StatusMethodNotAllowed 状态码
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	// 创建一些假数据，以便稍后填充表单中的字段
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n- Kobayashi Issa"
	expires := "7"

	// 
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet?id=%d", id), http.StatusSeeOther)
}