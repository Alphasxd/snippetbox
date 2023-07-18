package mysql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Alphasxd/snippetbox/pkg/models"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

// 创建新用户，将用户信息插入到数据库中
func (m *UserModel) Insert(name, email, password string) error {

	// 首先对密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	// 检查是否有重复的邮箱地址
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return models.ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

// 验证用户登录
func (m *UserModel) Authenticate(email, password string) (int, error) {

	var id int
	var hashedPassword []byte

	// 验证邮箱地址是否存在，还有用户是否还处于激活状态
	stmt := "SELECT id, hashed_password FROM users WHERE email = ? AND active = TRUE"
	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// 检验明文密码和哈希密码是否匹配
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// 如果上述验证都通过，则返回用户 id
	return id, nil
}

func (m *UserModel) Get(id int) (*models.User, error) {
	
	stmt := `SELECT id, name, email, created, active FROM users WHERE id = ?`
	row := m.DB.QueryRow(stmt, id)

	// 初始化一个指向 User struct 的指针
	u := &models.User{}

	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}

func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte
	row := m.DB.QueryRow("SELECT hashed_password FROM users WHERE id = ?", id)
	err := row.Scan(&currentHashedPassword)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return models.ErrInvalidCredentials
		} else {
			return err
		}
	}


	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt := "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)
	return err
}