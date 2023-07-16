package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Alphasxd/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

// 自定义一个类型，用于存储上下文密钥
type contextKey string

// 类型转换，将字符串转换为 contextKey 类型
const contextKeyIsAuthenticated = contextKey("isAuthenticated")

// 定义一个名为 application 的结构体
// 用于存储依赖注入的值，以及需要在整个应用程序中共享的状态信息
type application struct {
	infoLog       *log.Logger
	errorLog      *log.Logger
	session       *sessions.Session
	snippets      *mysql.SnippetModel
	users         *mysql.UserModel
	templateCache map[string]*template.Template
}

func main() {

	// 使用 flag 完成对服务端口的自定义设置，默认端口为 4000
	addr := flag.String("addr", ":4000", "HTTP newwork address")
	// 使用 flag 完成对 DSN 的自定义设置，默认值为 web:web@/snippetbox?parseTime=true
	dsn := flag.String("dsn", "web:web@/snippetbox?parseTime=true", "MySQL data source name")
	// 使用 flag 完成对 secret key 的自定义设置，默认值为 s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

	// 使用 flag.Parse() 解析命令行参数，必须在使用 flag 之后，访问任何命令行参数之前调用
	flag.Parse()

	// 定义两个 log.Logger 类型的日志记录器，一个用于记录信息日志，另一个用于记录错误日志
	// '|' 是按位或运算符，用于将标志参数连接起来，表示同时使用多个标志参数
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// 使用 openDB() 函数打开数据库连接
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	// 关闭数据库连接
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true // 设置 session cookie 为安全的，只能通过 HTTPS 来传输

	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		session: session,
		snippets: &mysql.SnippetModel{DB: db},
		users: &mysql.UserModel{DB: db},
		templateCache: templateCache,
	}

	// 初始化 tls.Config struct，设置服务器使用的 TLS 配置
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 使用 log.Println() 记录启动 web server 的日志信息
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err // 如果 Open() 失败，返回 nil 和错误
	}
	if err = db.Ping(); err != nil {
		return nil, err // 如果 Ping() 失败，返回 nil 和错误
	}
	return db, nil // 如果两个函数都成功，返回数据库连接和 nil
}