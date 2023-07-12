package mysql

import (
	"database/sql"
	"errors"

	"github.com/Alphasxd/snippetbox/pkg/models"
)

// 定义一个 SnippetModel 的 struct 类型，封装了一个 sql.DB 连接池
type SnippetModel struct {
	DB *sql.DB
}

// 向 snippets 表插入新的记录，返回新记录的 id 值
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	// 创建 SQL statement，用于向数据库插入新的记录，使用占位符代替参数
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// 使用 Exec() 方法执行 SQL statement，传入占位符参数
	// 返回一个 sql.Result 对象，包含一些关于这次操作结果的信息
	// 包括 LastInsertId() 和 RowsAffected() 方法
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// 使用 LastInsertId() 方法获取最后插入的记录的 id 值
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 将 id (int64) 转换为 int 类型，并返回
	return int(id), nil
}

// 通过 id 从 snippets 表中获取指定的记录
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	// 创建 SQL statement，用于从数据库中检索数据
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// 使用 QueryRow() 方法执行 SQL statement，传入占位符参数，返回一个指向该记录的指针
	row := m.DB.QueryRow(stmt, id)

	// 初始化一个指向 Snippet struct 的指针
	s := &models.Snippet{}

	// 如果查询没有匹配的记录，则 Scan() 方法会返回一个 sql.ErrNoRows 错误
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// 使用 errors.Is() 函数检查是否发生了 sql.ErrNoRows 错误
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	// 如果没有发生错误，则返回 Snippet struct 的指针
	return s, nil
}

// 获取 snippets 表中的最新 10 条记录，返回一个包含了这些记录的 []*Snippet 类型的切片
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return nil, nil
}