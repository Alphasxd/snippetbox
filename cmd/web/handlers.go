package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Alphasxd/snippetbox/pkg/forms"
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

//  home handler Get()
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

//  showSnippet handler Get()
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

// createSnippetForm handler Get()
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {

	// 使用 create.page.tmpl 模板渲染一个空白的表单
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// createSnippet handler Post()
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Snippet sucessfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}

// signupUserForm handler Get()
func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {

	// 使用 signup.page.tmpl 模板渲染一个空的表单
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// signupUser handler Post()
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {

	// 解析表单
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "flash", "Your signup was sucessful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

// loginUserForm handler Get()
func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {

	// 渲染一个空的表单
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// loginUser handler Post()
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {

	// 解析表单
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "authenticatedUserID", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

// logoutUser handler Post()
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {

	// 删除 session 中的 "authenticatedUserID" 键，以此来表示用户已经退出登录
	app.session.Remove(r, "authenticatedUserID")
	app.session.Put(r, "flash", "You've been logged out successfully!")

	// 回到主页
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "about.page.tmpl", nil)
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {

	userID := app.session.GetInt(r, "authenticatedUserID")

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "profile.page.tmpl", &templateData{
		User: user,
	})
}

func (app *application) changePasswordForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "password.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) changePassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("currentPassword", "newPassword", "newPasswordConfirmation")
	form.MinLength("newPassword", 10)
	if form.Get("newPassword") != form.Get("newPasswordConfirmation") {
		form.Errors.Add("newPasswordConfirmation", "Passwords do not match")
	}

	if !form.Valid() {
		app.render(w, r, "password.page.tmpl", &templateData{Form: form})
		return
	}

	userID := app.session.GetInt(r, "authenticatedUserID")

	err = app.users.ChangePassword(userID, form.Get("currentPassword"), form.Get("newPassword"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("currentPassword", "Current password is incorrect")
			app.render(w, r, "password.page.tmpl", &templateData{Form: form})
		} else if err != nil {
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "flash", "Your password has been updated!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}