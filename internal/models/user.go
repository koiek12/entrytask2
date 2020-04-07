package models

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type User struct {
	Id       string
	Password string
	Nickname string
	PicPath  string
}

// helper function to create statement for update in user table
func (u *User) updateStatement() string {
	var stmtStr bytes.Buffer
	i := 0
	if u.Password != "" {
		stmtStr.WriteString("password=")
		stmtStr.WriteString(fmt.Sprintf("'%s'", u.Password))
		i++
	}
	if u.Nickname != "" {
		if i > 1 {
			stmtStr.WriteString(",")
		}
		stmtStr.WriteString("nickname=")
		stmtStr.WriteString(fmt.Sprintf("'%s'", u.Nickname))
		i++
	}
	if u.PicPath != "" {
		if i > 1 {
			stmtStr.WriteString(",")
		}
		stmtStr.WriteString("pic_path=")
		stmtStr.WriteString(fmt.Sprintf("'%s'", u.PicPath))
		i++
	}
	return stmtStr.String()
}

// Get User information from DB
func GetUserById(db *sql.DB, id string) (*User, error) {
	var user User
	rows, err := db.Query(fmt.Sprintf("SELECT id, password, nickname, pic_path FROM USER WHERE id = '%s'", id))
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	err = rows.Scan(&user.Id, &user.Password, &user.Nickname, &user.PicPath)
	if err != nil {
		return nil, err
	}
	rows.Close()
	return &user, nil
}

// Update User information in DB
func SetUser(db *sql.DB, user *User) error {
	res, err := db.Exec(fmt.Sprintf("UPDATE USER SET %s WHERE id='%s'", user.updateStatement(), user.Id))
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

// Get User information from DB, compare it with password
func Authenticate(db *sql.DB, id, password string) (bool, error) {
	user, err := GetUserById(db, id)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, nil
	}
	hash := md5.Sum([]byte("salt#" + password))
	return hex.EncodeToString(hash[:]) == user.Password, nil
}
