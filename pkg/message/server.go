package message

import (
	"database/sql"
	"net"
	"os"

	"git.garena.com/youngiek.song/entry_task/internal/jwt"
	"git.garena.com/youngiek.song/entry_task/internal/models"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Server listens request from message.client.
type Server struct {
	listener    net.Listener                             // listener to accept new connection
	handlers    map[uint]func(*MsgStream, proto.Message) // pre-registered handlers for each request
	db          *sql.DB                                  // database connection to user DB
	tokenIssuer *jwt.TokenIssuer                         // Generate and Authenticate JWT Token with secret Key
	logger      *zap.Logger                              // for log
	host, port  string                                   // listen host and port
}

// NewServer create new instance of server.
func NewServer(host, port string, db *sql.DB, tokenIssuer *jwt.TokenIssuer, logger *zap.Logger) *Server {
	// initialize listen socket
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		logger.Fatal("Error opening listen socket")
		os.Exit(1)
	}

	server := &Server{
		host:        host,
		port:        port,
		listener:    listener,
		handlers:    make(map[uint]func(*MsgStream, proto.Message)),
		db:          db,
		tokenIssuer: tokenIssuer,
		logger:      logger,
	}
	// register handler for each message
	server.registerHandler(&HealthcheckMessage{}, server.healthCheck)
	server.registerHandler(&LoginRequest{}, server.login)
	server.registerHandler(&GetUserInfoRequest{}, server.getUserInfo)
	server.registerHandler(&EditUserInfoRequest{}, server.editUserInfo)
	server.registerHandler(&AuthRequest{}, server.authenticate)
	return server
}

// registerHandler function to server's handler
func (server *Server) registerHandler(msg proto.Message, handler func(*MsgStream, proto.Message)) error {
	msgNum, err := getMsgNum(msg)
	if err != nil {
		return err
	}
	server.handlers[msgNum] = handler
	return nil
}

// getHandler map message to it's corresponding handler function.
func (server *Server) getHandler(msg proto.Message) (func(*MsgStream, proto.Message), error) {
	msgNum, err := getMsgNum(msg)
	if err != nil {
		return nil, err
	}
	return server.handlers[msgNum], nil
}

// Run start the server listening to tcp request.
func (server *Server) Run() {
	server.logger.Info("Backend Server has started, Listening on " + net.JoinHostPort(server.host, server.port) + "...")
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			server.logger.Error("Error accepting connection", zap.String("error", err.Error()))
			break
		}
		go server.handleRequest(conn)
	}
}

// handleRequest process request and send response to client.
func (server *Server) handleRequest(conn net.Conn) {
	stream, _ := NewMsgStream(conn, 60)
	defer server.logger.Info("close connection", zap.String("remote", stream.RemoteAddr()))
	defer stream.Close()
	for {
		//wait for next request
		msg, err := stream.ReadMsg()
		if err != nil {
			if terr, ok := err.(net.Error); ok && terr.Timeout() {
				server.logger.Info("Connection timeout waiting for new request", zap.String("remote", stream.RemoteAddr()))
			} else {
				server.logger.Error("Error receiving request", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
			}
			break
		}
		handler, err := server.getHandler(msg)
		if err != nil {
			server.logger.Error("Not handler registered message", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
			break
		}
		handler(stream, msg)
	}
}

// healthCheck handles healthCheck message from client. It is just for check health of server.
func (server *Server) healthCheck(stream *MsgStream, r proto.Message) {
	stream.WriteMsg(&HealthcheckMessage{})
}

// login handles login request. Compare password of user with db's data.
// On success, response with generated jwt token and error code 0.
// On fail, response with empty token and positive error code.
func (server *Server) login(stream *MsgStream, r proto.Message) {
	req := r.(*LoginRequest)
	id := req.Id
	password := req.Password
	valid, err := models.Authenticate(server.db, id, password)
	if err != nil {
		server.logger.Error("Error authenticating id/password", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&LoginResponse{
			Response: &Response{Code: 2},
		})
		return
	}
	if !valid {
		server.logger.Warn("invalid Id/password", zap.String("remote", stream.RemoteAddr()), zap.String("id", id), zap.String("password", password))
		stream.WriteMsg(&LoginResponse{
			Response: &Response{Code: 1},
		})
		return
	}
	msg := &LoginResponse{
		Response: &Response{Code: 0},
		Token:    server.tokenIssuer.GenerateToken(id),
	}
	stream.WriteMsg(msg)
	server.logger.Info("Handled login request", zap.String("remote", stream.RemoteAddr()), zap.String("id", id))
}

// getUserInfo check client's priviliege by JWT token and get user's information from DB.
// On success, response with user data and error code 0.
// On fail, response with user data and positive error code.
func (server *Server) getUserInfo(stream *MsgStream, r proto.Message) {
	req := r.(*GetUserInfoRequest)
	id, err := server.tokenIssuer.AuthenticateToken(req.Token)
	if err != nil {
		server.logger.Error("Token authentication failed", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 3},
		})
		return
	}
	if id == "" {
		server.logger.Warn("Invalid token", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 1},
		})
		return
	}
	user, err := models.GetUserById(server.db, id)
	if err != nil {
		server.logger.Error("Error on DB", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 2},
		})
		return
	}
	if user == nil {
		server.logger.Info("No such user", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&GetUserInfoResponse{
			Response: &Response{Code: 3},
		})
		return
	}
	stream.WriteMsg(&GetUserInfoResponse{
		Response: &Response{Code: 0},
		User: &User{
			Id:       user.Id,
			Nickname: user.Nickname,
			PicPath:  user.PicPath,
		},
	})
	server.logger.Info("Handled GetUserInfo request", zap.String("remote", stream.RemoteAddr()), zap.String("id", id))
}

// editUserInfo check client's priviliege by JWT token and edit user's information from DB.
// On success, response with error code 0.
// On fail, response with positive error code.
func (server *Server) editUserInfo(stream *MsgStream, r proto.Message) {
	req := r.(*EditUserInfoRequest)
	id, err := server.tokenIssuer.AuthenticateToken(req.Token)
	if err != nil {
		server.logger.Error("Token authentication failed", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&Response{Code: 3})
		return
	}
	if id == "" {
		server.logger.Warn("Invalid token", zap.String("remote", stream.RemoteAddr()), zap.String("token", req.Token))
		stream.WriteMsg(&Response{Code: 1})
		return
	}
	err = models.SetUser(server.db, &models.User{
		Id:       req.User.Id,
		Nickname: req.User.Nickname,
		PicPath:  req.User.PicPath,
	})
	if err != nil {
		server.logger.Error("Error on DB", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&Response{Code: 2})
		return
	}
	stream.WriteMsg(&Response{Code: 0})
	server.logger.Info("Handled EditUserInfo request", zap.String("remote", stream.RemoteAddr()), zap.String("id", id), zap.String("body", req.User.String()))
}

// getUserInfo check client's priviliege by JWT token.
// On success, response with error code 0.
// On fail, response with positive error code.
func (server *Server) authenticate(stream *MsgStream, r proto.Message) {
	req := r.(*AuthRequest)
	id, err := server.tokenIssuer.AuthenticateToken(req.Token)
	if err != nil {
		server.logger.Error("Token authentication failed", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&Response{Code: 1})
		return
	}
	if id == "" {
		server.logger.Info("Invalid token", zap.String("remote", stream.RemoteAddr()), zap.String("error", err.Error()))
		stream.WriteMsg(&Response{Code: 1})
		return
	}
	stream.WriteMsg(&Response{Code: 0})
	server.logger.Info("Handled Authenticate request", zap.String("remote", stream.RemoteAddr()), zap.String("id", id))
}
