package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/Alphasxd/snippetbox/pkg/forms"
	"github.com/Alphasxd/snippetbox/pkg/models"
	"github.com/Alphasxd/snippetbox/ui"
)

// templateData 用于存储应用程序中的动态数据，这些数据将传递到 HTML 模板中
type templateData struct {
	CSRFToken       string
	CurrentYear     int
	Flash           string
	Form            *forms.Form
	IsAuthenticated bool
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	User            *models.User
}

// humanDate 将时间对象格式化为人类可读的字符串
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

// newTemplateCache 用于创建一个新的模板缓存
func newTemplateCache() (map[string]*template.Template, error) {
	// 初始化一个新的模板缓存 map
	cache := map[string]*template.Template{}

	// 使用 filepath.Glob 函数获取模板目录下所有以 ".page.tmpl" 结尾的模板文件
	pages, err := fs.Glob(ui.Files, "html/*.page.tmpl")
	if err != nil {
		return nil, err
	}

	// 一个一个地循环这些文件
	for _, page := range pages {
		name := filepath.Base(page)

		// 加载模板文件到一个 template.Template 对象中
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, page)
		if err != nil {
			return nil, err
		}

		// 加载 layout 文件到 template.Template 对象中
		ts, err = ts.ParseFS(ui.Files, "html/*.layout.tmpl")
		if err != nil {
			return nil, err
		}

		// 加载 partial 文件到 template.Template 对象中
		ts, err = ts.ParseFS(ui.Files, "html/*.partial.tmpl")
		if err != nil {
			return nil, err
		}

		// 将 template.Template 对象添加到缓存 map 中，键是文件名
		cache[name] = ts
	}

	// 返回模板缓存 map
	return cache, nil
}
