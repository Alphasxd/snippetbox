package main

import (
	"errors"
	"fmt"
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

//  home handler
func (app *application) home(w http.ResponseWriter, r *http.Request) {

	// 通过调用 SnippetModel 的 Latest() 方法来获取最新的 10 个snippet
	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// 使用 render() helper 方法来渲染模板
	app.render(w, r, "home.page.tmpl", &templateData{
		Snippets: s,
	})
}

//  showSnippet handler
func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {

	// 使用 r.URL.Query().Get() 方法获取 "id" 查询字符串参数的值
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))

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

	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})

}

// createSnippetForm handler
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "create.page.tmpl", nil)
}

// createSnippet handler
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {

	// 创建一些假数据，以便稍后填充表单中的字段
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n- Kobayashi Issa"
	expires := "7"

	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}