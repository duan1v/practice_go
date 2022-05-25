package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
)

type User struct {
	Name string
	Addr string
	Conn net.Conn
	Ch   chan string // 消息通道
	// done   chan bool   // 是否已结束聊天
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		Ch:   make(chan string),
		// done:   make(chan bool),
		server: server,
	}
	go user.ListenMessage()
	return user
}

func (user *User) Online() {
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	user.server.CollectMessage(user, "已上线！")
}

func (user *User) ValidateAndGoexit() {
	_, err := user.server.OnlineMap[user.Name]
	if !err {
		runtime.Goexit()
	}
}

func (user *User) Offline() {
	user.ValidateAndGoexit()
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	user.server.CollectMessage(user, "已下线！")
	// 释放用户资源
	close(user.Ch)
	// user.done <- true
}

func (user *User) SendMessage(msg string) {
	// 通过conn向客户端输出写入消息
	_, err := user.Conn.Write([]byte(msg))
	if err != nil && err != io.EOF {
		fmt.Println("conn write err ", err, ";message:", msg)
		return
	}
}

func (user *User) Domessage(msg string) {
	if strings.TrimSpace(msg) == "" {
		user.SendMessage("不可以输入空内容")
		return
	} else if msg == "who" {
		// 用户输入who，可查询在线用户
		user.server.mapLock.Lock()
		for _, onlineUser := range user.server.OnlineMap {
			user.SendMessage(fmt.Sprintf("[%s]%s:%s\n", onlineUser.Addr, onlineUser.Name, "在线"))
		}
		user.server.mapLock.Unlock()
	} else if ml := len(msg); ml > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, err := user.server.OnlineMap[newName]
		if err {
			user.SendMessage(fmt.Sprintf("%s,这个名字已被占用\n", newName))
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMessage(fmt.Sprintf("你好,%s,名字修改成功\n", newName))
		}
	} else if len(msg) > 3 && msg[:3] == "to|" {
		body := strings.Split(msg, "|")
		if len(body) != 3 {
			user.SendMessage("发送信息格式错误;请以'to|张三|你好'的格式发送。\n")
			return
		}
		name := body[1]
		if name == user.Name {
			user.SendMessage("不可以给自己发送消息。\n")
			return
		}
		content := body[2]
		if strings.TrimSpace(content) == "" {
			user.SendMessage("发送内容不能为空。\n")
			return
		}
		toUser, err := user.server.OnlineMap[name]
		if !err {
			user.SendMessage(fmt.Sprintf("%s,这个用户不在线\n", name))
		} else {
			toUser.SendMessage(fmt.Sprintf("%s,对您说:%s\n", user.Name, content))
		}
	} else {
		// 是让服务端收集信息，利用广播向所有客户端发送信息
		user.server.CollectMessage(user, msg)
	}
}

func (user *User) ListenMessage() {
	// 6、除了服务器端主动动作的死循环不需要考虑条件外，涉及到客户端的死循环，必须注意结束条件
	for {
		// // 4、写法一、给user添加结束信号通道
		// select {
		// case <-user.done:
		// 	// 关闭连接
		// 	user.Conn.Close()
		// 	close(user.done)
		// 	runtime.Goexit()
		// default:
		// 	// 1、从客户端通道获取信息，写入到conn中
		// 	// 2、如果channel关闭，则此处不再阻塞！！！！！！！！！！！向已关闭的channel读取，会造成broken pipe的错误
		// 	// 3、也就是说，如果继续for循环，这个会一直调用下面那个写到客户端的方法
		// 	msg := <-user.Ch
		// 	user.SendMessage(msg)
		// }
		// 7、上面的写法还是会有几率产生broken pipe；猜测是在Offline方法中
		// user.server.CollectMessage(user, "已下线！") 导致的 BroadCast()方法 在 close(user.Ch) 之后执行了
		// 5、写法二、直接断言
		msg, ok := <-user.Ch
		if !ok {
			// 关闭连接
			user.Conn.Close()
			return
		}
		user.SendMessage(msg)
	}
}
