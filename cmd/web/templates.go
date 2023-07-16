package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/Alphasxd/snippetbox/pkg/forms"
	"github.com/Alphasxd/snippetbox/pkg/models"
)

type templateData struct {
	CurrentYear int
	Flash           string
	Form            *forms.Form
	IsAuthenticated bool
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap {
	"humanDate": humanDate,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// 初始化一个新的模板缓存 map
	cache := map[string]*template.Template{}

	// 使用 filepath.Glob 函数获取模板目录下所有以 ".page.tmpl" 结尾的模板文件
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// 一个一个地循环这些文件
	for _, page := range pages {
		name := filepath.Base(page)

		// 加载模板文件到一个 template.Template 对象中
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// 加载 layout 文件到 template.Template 对象中
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// 加载 partial 文件到 template.Template 对象中
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// 将 template.Template 对象添加到缓存 map 中，键是文件名
		cache[name] = ts 
	}

	// 返回模板缓存 map
	return cache, nil
}

