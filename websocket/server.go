package websocket

import (
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/saikitanabe/go-engine.io/message"
	"github.com/saikitanabe/go-engine.io/parser"
	"github.com/saikitanabe/go-engine.io/transport"
)

type Server struct {
	callback transport.Callback
	conn     *websocket.Conn
}

func NewServer(w http.ResponseWriter, r *http.Request, callback transport.Callback) (transport.Server, error) {
	// conn, err := websocket.Upgrade(w, r, nil, 10240, 10240)

	// use Gorilla's default websocket check origin to prevent access from other domains
	conn, err := upgradeWebsocket(w, r, nil, 10240, 10240)
	if err != nil {
		return nil, err
	}

	ret := &Server{
		callback: callback,
		conn:     conn,
	}

	go ret.serveHTTP(w, r)

	return ret, nil
}

func upgradeWebsocket(w http.ResponseWriter, r *http.Request, responseHeader http.Header, readBufSize, writeBufSize int) (*websocket.Conn, error) {
	u := websocket.Upgrader{ReadBufferSize: readBufSize, WriteBufferSize: writeBufSize}
	u.Error = func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		// don't return errors to maintain backwards compatibility
	}

	// Don't set this; do not allow connect from any domain
	// u.CheckOrigin = func(r *http.Request) bool {
	// 	// allow all connections by default
	// 	return true
	// }

	// fmt.Println("upgrader...")
	return u.Upgrade(w, r, responseHeader)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (s *Server) NextWriter(msgType message.MessageType, packetType parser.PacketType) (io.WriteCloser, error) {
	wsType, newEncoder := websocket.TextMessage, parser.NewStringEncoder
	if msgType == message.MessageBinary {
		wsType, newEncoder = websocket.BinaryMessage, parser.NewBinaryEncoder
	}

	w, err := s.conn.NextWriter(wsType)
	if err != nil {
		return nil, err
	}
	ret, err := newEncoder(w, packetType)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	defer s.callback.OnClose(s)

	for {
		t, r, err := s.conn.NextReader()
		if err != nil {
			s.conn.Close()
			return
		}

		switch t {
		case websocket.TextMessage:
			fallthrough
		case websocket.BinaryMessage:
			decoder, err := parser.NewDecoder(r)
			if err != nil {
				return
			}
			s.callback.OnPacket(decoder)
			decoder.Close()
		}
	}
}
