package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"git.garena.com/youngiek.song/entry_task/internal/cache"
	"git.garena.com/youngiek.song/entry_task/internal/jwt"
	"git.garena.com/youngiek.song/entry_task/pkg/message"
	"go.uber.org/zap"
)

// UserController provides handler functions for http server. UserController is also able to access injected dependecies.
type UserController struct {
	client  *message.Client
	cache   *cache.UserCache
	logger  *zap.Logger
	docRoot string
}

// NewUserController create new instance of user controller with injected dependencies.
func NewUserController(client *message.Client, cache *cache.UserCache, logger *zap.Logger, docRoot string) *UserController {
	return &UserController{
		client:  client,
		cache:   cache,
		logger:  logger,
		docRoot: docRoot,
	}
}

// LoginPage shows login page to user.
func (controller *UserController) LoginPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(controller.docRoot + "/template/login.html")
	t.Execute(w, nil)
}

// Login tries login with id/password. Authenticate user's id/password by sending login request to backend TCP server.
// If successful, issue JWT access token to user's cookie and redirect to main page.
func (controller *UserController) Login(w http.ResponseWriter, r *http.Request) {
	id := r.PostFormValue("id")
	passwd := r.PostFormValue("pwd")

	token, err := controller.client.Login(id, passwd)
	if err != nil {
		switch err.(type) {
		case message.AuthError:
			controller.logger.Warn("Wrong Id/Password", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("id", id), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Wrong ID/Password.", err)
		default:
			controller.logger.Error("Fail communicating backend server", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Server error.", err)
		}
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: token, Path: "/"})
	http.Redirect(w, r, "/main", 302)
	controller.logger.Info("Login request", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path))
}

// Main shows user main page which contains user's information.
// User should have JWT access token as cookie to retrieve the information from backend TCP server.
// After retrieving user info from backend server, it draws main page with the information.
func (controller *UserController) Main(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		controller.logger.Info("No access token", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Can't access this page. You don't have access token.", err)
		return
	}
	id, err := jwt.GetIDFromToken(tokenCookie.Value)
	if err != nil {
		controller.logger.Warn("No ID claim in access token", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Can't access this page. Invalid access token.", err)
		return
	}
	user, err := controller.cache.GetUserInfo(id)
	if err == nil && user != nil {
		err = controller.client.Authenticate(tokenCookie.Value)
		if err != nil {
			switch err.(type) {
			case message.AuthError:
				http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1})
				controller.logger.Warn("Access token authentication fail", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintln(w, "Server Error.", err)
			default:
				controller.logger.Error("Fail communicating backend server", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Server Error.", err)
			}
			return
		}
	} else {
		user, err = controller.client.GetUserInfo(tokenCookie.Value)
		if err != nil {
			controller.logger.Error("Fail communicating backend server", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Server error", err)
			return
		}
		controller.cache.SetUserInfo(user)
	}
	t, _ := template.ParseFiles(controller.docRoot + "/template/main.html")
	t.Execute(w, user)
	controller.logger.Info("request success", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value))
}

// EditUserInfo modify user's information.
// User should have JWT access token as cookie to retrieve the information from backend TCP server.
// After successfully modifying user info from backend server, it redirect to main page.
func (controller *UserController) EditUserInfo(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		controller.logger.Error("No access token", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Can't access this page. You don't have access token.", err)
		return
	}
	id, err := jwt.GetIDFromToken(tokenCookie.Value)
	if err != nil {
		controller.logger.Warn("No ID claim in access token", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Can't access this page. Invalid access token.", err)
		return
	}
	nickname := r.PostFormValue("nickname")

	err = controller.client.EditUserInfo(tokenCookie.Value, &message.User{
		Id:       id,
		Nickname: nickname,
	})
	if err != nil {
		switch err.(type) {
		case message.AuthError:
			controller.logger.Warn("Access token authentication fail", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Server Error.", err)
		default:
			controller.logger.Error("Fail communicating backend server", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Server error", err)
		}
		return
	}
	err = controller.cache.DelUserInfo(id)
	if err != nil {
		controller.logger.Error("Delete user cache fail", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("id", id), zap.String("error", err.Error()))
	}
	http.Redirect(w, r, "/main", 302)
	controller.logger.Info("request success", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value))
}

// UploadPhoto uploads user's profile picture.
// User should have JWT access token as cookie to authenticate your access priviligies from TCP backend server.
// After successfully modifying picture, it redirect to main page.
func (controller *UserController) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("access_token")
	if err != nil {
		controller.logger.Error("No access token", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Can't access this page. You don't have access token.", err)
		return
	}
	mulFile, _, err := r.FormFile("picFile")
	defer mulFile.Close()
	if err != nil {
		controller.logger.Error("Error reading picture file", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "File error.", err)
		return
	}

	user, err := controller.client.GetUserInfo(tokenCookie.Value)
	if err != nil {
		switch err.(type) {
		case message.AuthError:
			controller.logger.Warn("Access token authentication fail", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(w, "Server Error.", err)
		default:
			controller.logger.Error("Fail communicating backend server", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value), zap.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Server error.", err)
		}
		return
	}

	outFile, err := os.OpenFile(controller.docRoot+"/static/"+user.PicPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		controller.logger.Error("Error opening picture file", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Server error.", err)
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, mulFile)
	if err != nil {
		controller.logger.Error("Error copying picture file", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Server error.", err)
		return
	}
	http.Redirect(w, r, "/main", 302)
	controller.logger.Info("UploadPhoto request", zap.String("remote", r.RemoteAddr), zap.String("path", r.URL.Path), zap.String("token", tokenCookie.Value))
}
