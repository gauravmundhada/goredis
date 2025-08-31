package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Message struct {
	data []byte
	peer *Peer
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message
	kv        *KV
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKV(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	slog.Info("server running", "addr", s.ListenAddr)

	return s.acceptLoop()
}

func (s *Server) handleMessage(msg Message) error {
	cmd, err := parseCommand(string(msg.data))
	//fmt.Printf("cmd type: %T, value: %#v\n", cmd, cmd)

	// switch v := cmd.(type) {
	// case SetCommand:
	// 	return s.kv.Set(v.key, v.val)
	// case GetCommand:
	// 	val, ok := s.kv.Get(v.key)
	// 	if !ok {
	// 		return fmt.Errorf("key not found")
	// 	}
	// 	fmt.Println(string(val))
	// 	_, err := msg.peer.Send(val)
	// 	if err != nil {
	// 		slog.Error("peer send error", "err", err)
	// 	}
	// }

	if err != nil {
		return err
	}
	switch v := cmd.(type) {
	case SetCommand:
		//slog.Info("SET command triggered", "key", cmd.(SetCommand).key, "val", cmd.(SetCommand).val)
		return s.kv.Set(v.key, v.val)
	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}
	}
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			fmt.Println("Received message:", string(msg.data))
			if err := s.handleMessage(msg); err != nil {
				slog.Error("handle raw message error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true

		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer

	slog.Info("new peer connected", "remoteAddr", conn.RemoteAddr())
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read loop error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}

func main() {
	//go func() {
	server := NewServer(Config{})
	log.Fatal(server.Start())
	//}()
	// time.Sleep(time.Second) // wait for server to start

	// client := client.New("localhost:5001")
	// for i := 0; i < 10; i++ {
	// 	if err := client.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	v, err := client.Get(context.TODO(), fmt.Sprintf("foo_%d", i))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(v)
	// }

	// time.Sleep(time.Second)
	// fmt.Println(server.kv.data)
}
