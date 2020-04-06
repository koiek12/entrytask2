package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"git.garena.com/youngiek.song/entry_task/internal/cache"
	"git.garena.com/youngiek.song/entry_task/internal/jwt"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
)

var client = message.NewClient()
var userCache = cache.NewUserCache()

func Login(w http.ResponseWriter, r *http.Request) {
	logger.Instance.Info(r.RemoteAddr + " Login request")
	id := r.PostFormValue("id")
	passwd := r.PostFormValue("pwd")
	c, err := r.Cookie("access_token")
	if err == nil && c.Value != "" {
		logger.Instance.Info("Already have acces token.")
		http.Redirect(w, r, "/main", 302)
		return
	}
	token, err := client.Login(id, passwd)
	if err != nil {
		logger.Instance.Error("Fail communicating backend server.")
		fmt.Fprintln(w, "Failed to Login.", err)
		return
	}
	cookie := http.Cookie{Name: "access_token", Value: token, Path: "/"}
	http.SetCookie(w, &cookie)
	logger.Instance.Info(r.RemoteAddr + " Login response")
	http.Redirect(w, r, "/main", 302)
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	logger.Instance.Info(r.RemoteAddr + " GetUserInfo request")
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		logger.Instance.Info("No access token")
		fmt.Fprintln(w, "Can't access this page. You don't have access token.", err)
		return
	}
	id, err := jwt.GetIdFromToken(tokenCookie.Value)
	if err != nil {
		logger.Instance.Warn("No ID claim in access token")
		fmt.Fprintln(w, "Can't access this page. Invalid access token.", err)
		return
	}
	user, err := userCache.GetUserInfo(id)
	if err == nil && user != nil {
		res, err := client.Authenticate(tokenCookie.Value)
		if err != nil {
			logger.Instance.Error("Fail communicating backend server.")
			fmt.Fprintln(w, "Failed to authenticate access token.", err)
			return
		}
		if res == false {
			logger.Instance.Warn("Invalid access token.")
			fmt.Fprintln(w, "Can't access this page. Invalid access token.")
			return
		}
	} else {
		user, err = client.GetUserInfo(tokenCookie.Value)
		if err != nil {
			logger.Instance.Error("Fail communicating backend server.")
			fmt.Fprintln(w, "Failed to get userInfo.", err)
			return
		}
		userCache.SetUserInfo(user)
	}
	t, _ := template.ParseFiles("../../web/template/main.html")
	t.Execute(w, user)
	logger.Instance.Info(r.RemoteAddr + " GetuserInfo response")
}

func EditUserInfo(w http.ResponseWriter, r *http.Request) {
	logger.Instance.Info(r.RemoteAddr + " EditUserInfo request")
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		logger.Instance.Error("No access token")
		fmt.Fprintln(w, "Can't access this page. You don't have access token.", err)
		return
	}
	id, _ := jwt.GetIdFromToken(tokenCookie.Value)
	nickname := r.PostFormValue("nickname")

	err = client.EditUserInfo(tokenCookie.Value, &message.User{
		Id:       id,
		Nickname: nickname,
	})
	if err != nil {
		logger.Instance.Error("Fail communicating backend server.")
		fmt.Fprintln(w, "Failed to edit userInfo.", err)
		return
	}
	err = userCache.DelUserInfo(id)
	if err != nil {
		logger.Instance.Error("Delete cache fail")
	}
	http.Redirect(w, r, "/main", 302)
	logger.Instance.Info(r.RemoteAddr + " EditUserInfo request")
}

func UploadPhoto(w http.ResponseWriter, r *http.Request) {
	logger.Instance.Info(r.RemoteAddr + " UploadPhoto request")
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		logger.Instance.Error("No access token")
		fmt.Fprintln(w, "Can't access this page. You don't have access token.", err)
		http.Redirect(w, r, "/", 302)
		return
	}
	mulFile, _, err := r.FormFile("picFile")
	defer mulFile.Close()
	if err != nil {
		logger.Instance.Error("Error reading file")
		fmt.Fprintln(w, "File error.", err)
		return
	}

	user, err := client.GetUserInfo(tokenCookie.Value)
	if err != nil {
		logger.Instance.Error("Fail communicating backend server")
		fmt.Fprintln(w, "Can't access this page. Invalid access token.", err)
		return
	}

	outFile, _ := os.OpenFile("../../web/static/"+user.PicPath, os.O_WRONLY|os.O_CREATE, 0666)
	defer outFile.Close()
	io.Copy(outFile, mulFile)
	http.Redirect(w, r, "/main", 302)
	logger.Instance.Info(r.RemoteAddr + " UploadPhoto response")
}
