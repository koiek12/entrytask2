package message

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	reflect "reflect"

	"git.garena.com/youngiek.song/entry_task/internal/jwt"
	"git.garena.com/youngiek.song/entry_task/internal/logger"
	"git.garena.com/youngiek.song/entry_task/internal/models"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	listener    net.Listener
	db          *sql.DB
	msgHandlers map[string]func(*MsgStream, *sql.DB, proto.Message)
}

func NewServer(host, port string) *Server {
	// initialize listen socket
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		logger.Instance.Fatal("Error opening listen socket")
		os.Exit(1)
	}
	// initialize database connection
	db, err := sql.Open("mysql", "song:abcd@/entry_task")
	if err != nil {
		logger.Instance.Fatal("Cannot connect to DB " + err.Error())
		os.Exit(1)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	// check database alive
	err = db.Ping()
	if err != nil {
		logger.Instance.Fatal("Cannot connect to DB " + err.Error())
		os.Exit(1)
	}

	msgHandlers := make(map[string]func(*MsgStream, *sql.DB, proto.Message))
	msgHandlers[reflect.TypeOf(&HealthcheckMessage{}).String()] = handleHealthCheck
	msgHandlers[reflect.TypeOf(&LoginRequest{}).String()] = handleLogin
	msgHandlers[reflect.TypeOf(&GetUserInfoRequest{}).String()] = handleGetUserInfo
	msgHandlers[reflect.TypeOf(&EditUserInfoRequest{}).String()] = handleEditUserInfo
	msgHandlers[reflect.TypeOf(&AuthRequest{}).String()] = handleAuthRequest
	return &Server{
		listener:    listener,
		db:          db,
		msgHandlers: msgHandlers,
	}
}

func (s *Server) getHandler(msg proto.Message) func(*MsgStream, *sql.DB, proto.Message) {
	return s.msgHandlers[reflect.TypeOf(msg).String()]
}

func (s *Server) Run() {
	for {
		logger.Instance.Info("Backend Server has started, Listening on port 3233...")
		conn, err := s.listener.Accept()
		if err != nil {
			logger.Instance.Error("Error accepting connection")
			break
		}
		go s.handleRequest(conn)
	}
}

func (s *Server) handleRequest(conn net.Conn) {
	stream, _ := NewMsgStream(conn, 60)
	defer logger.Instance.Info("close connection")
	defer stream.Close()
	for {
		msg, err := stream.ReadMsg()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				logger.Instance.Info("Timeout waiting for new message")
			} else {
				logger.Instance.Error("Error receiving message")
			}
			break
		}
		handler := s.getHandler(msg)
		handler(stream, s.db, msg)
	}
}

func handleHealthCheck(st *MsgStream, db *sql.DB, r proto.Message) {
	st.WriteMsg(&HealthcheckMessage{})
}

func handleLogin(st *MsgStream, db *sql.DB, r proto.Message) {
	req := r.(*LoginRequest)
	id := req.Id
	password := req.Password
	valid, err := models.Authenticate(db, id, password)
	if err != nil {
		logger.Instance.Error("Error on DB." + err.Error())
		st.WriteMsg(&LoginResponse{
			Response: &Response{Code: 2},
		})
		return
	}
	if !valid {
		logger.Instance.Debug("Wrong password.")
		st.WriteMsg(&LoginResponse{
			Response: &Response{Code: 1},
		})
		return
	}
	token := jwt.GenerateToken(id)

	msg := &LoginResponse{
		Response: &Response{Code: 0},
		Token:    token,
	}
	st.WriteMsg(msg)
}

func handleGetUserInfo(st *MsgStream, db *sql.DB, r proto.Message) {
	req := r.(*GetUserInfoRequest)
	id, err := jwt.AuthenticateToken(req.Token)
	if err != nil {
		logger.Instance.Debug("Token authentication failed" + err.Error())
		st.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 1},
		})
		return
	}
	user, err := models.GetUserById(db, id)
	if err != nil {
		logger.Instance.Error("Error on DB." + err.Error())
		st.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 2},
		})
		return
	}
	if user == nil {
		logger.Instance.Debug("No such user")
		st.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 3},
		})
		return
	}
	st.WriteMsg(&GetUserInfoResponse{
		Response: &Response{Code: 0},
		User: &User{
			Id:       user.Id,
			Nickname: user.Nickname,
			PicPath:  user.PicPath,
		},
	})
}

func handleEditUserInfo(st *MsgStream, db *sql.DB, r proto.Message) {
	req := r.(*EditUserInfoRequest)
	_, err := jwt.AuthenticateToken(req.Token)
	if err != nil {
		fmt.Println("Authenticate token failed. Invalid token.")
		st.WriteMsg(&Response{Code: 1})
		return
	}
	err = models.SetUser(db, &models.User{
		Id:       req.User.Id,
		Nickname: req.User.Nickname,
		PicPath:  req.User.PicPath,
	})
	if err != nil {
		fmt.Println("Failed to access DB.")
		st.WriteMsg(&Response{Code: 2})
		return
	}
	st.WriteMsg(&Response{Code: 0})
}

func handleAuthRequest(st *MsgStream, db *sql.DB, r proto.Message) {
	req := r.(*AuthRequest)
	_, err := jwt.AuthenticateToken(req.Token)
	if err != nil {
		fmt.Println("Authenticate token failed. Invalid token.")
		st.WriteMsg(&Response{Code: 1})
		return
	}
	st.WriteMsg(&Response{Code: 0})
}
