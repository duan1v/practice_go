package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	Protocol  string
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Ch        chan string
}

func NewServer(ip string, port int, protocol string) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		Protocol:  protocol,
		OnlineMap: make(map[string]*User),
		Ch:        make(chan string),
	}

	return server
}

func (server *Server) BroadCast() {
	for {
		msg := <-server.Ch
		server.mapLock.Lock()
		for _, user := range server.OnlineMap {
			user.Ch <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) HintUser(user *User, msg string) {
	user.Ch <- msg
}

// 服务器收集信息
func (server *Server) CollectMessage(user *User, msg string) {
	fmtMsg := fmt.Sprintf("[%s]%s:%s\n", user.Addr, user.Name, msg)
	server.Ch <- fmtMsg
}

func (server *Server) Handler(conn net.Conn) {
	fmt.Println("创建连接成功！")
	// 有个用户连接了
	user := NewUser(conn, server)
	user.Online()
	isLive := make(chan bool)
	isRead := make(chan bool)
	go func() {
		buf := make([]byte, 4096)
		msg := ""
		for {
			select {
			case <-isRead: // 结束当前协程
				close(isLive)
				close(isRead)
				runtime.Goexit() // os.Exit(1) 这个会将整个服务端退出，不能使用
			default:
				// conn读取客户端信息
				n, err := conn.Read(buf)
				// 读取到ctrl+c之类的断开消息
				if n == 0 {
					user.Offline()
					return
				}

				if err != nil && err != io.EOF {
					fmt.Println("conn read err ", err)
					return
				}

				// 去除最后一个换行字节\n(按回车输入时的换行)
				if buf[n-1] == '\n' && (strings.TrimSpace(msg) != "" || len(buf) > 1) {
					user.Domessage(msg + string(buf[:n-1]))
					msg = ""
				} else {
					msg += string(buf)
				}
				isLive <- true
			}
		}
	}()
	for {
		select {
		case _, ok := <-isLive: // 当前用户是活跃的，为了激活select，更新下面的计时器
			if !ok { // 表示isLive已经close
				return // 不结束的话，还是会一直执行；由于select的随机性，也可能走下面一个
			}
		case <-time.After(time.Second * 60 * 5):
			user.ValidateAndGoexit()
			// 通知客户端
			user.SendMessage("已被踢出群聊")
			user.Offline()
			runtime.Goexit()
			// 结束对这个客户上面的读取输入操作，否则上面的协程会一直存在
			isRead <- true
		}
	}
}

func (server *Server) Start() {
	// 构件socket
	listener, err := net.Listen(server.Protocol, fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net listener err ", err)
	}
	defer listener.Close()

	go server.BroadCast()

	for {
		// 构件连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("net listener accept err ", err)
			continue
		}
		// 有客户端搭理这个服务器，就去派个goroutine处理
		go server.Handler(conn)
	}
}
