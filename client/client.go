package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Client struct {
	Host string
	Port int
	Name string
	Conn net.Conn
	flag int
}

func NewClient(host string, port int) *Client {
	client := &Client{
		Host: host,
		Port: port,
		flag: 999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port))

	if err != nil {
		fmt.Println("net dial err: ", err)
		os.Exit(1)
	}

	client.Conn = conn

	return client
}

var (
	serverHost string
	serverPort int
)

func init() {
	flag.StringVar(&serverHost, "host", "127.0.0.1", "设置服务器主机地址")
	flag.IntVar(&serverPort, "port", 8081, "设置服务器端口")
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println(strings.Repeat("=", 20))
	fmt.Println("1.世界频道")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.获取当前在线用户")
	fmt.Println("0.退出")
	fmt.Println(strings.Repeat("=", 20))

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 4 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>请输入合法范围内的数字<<<<<<<")
		return false
	}
}

func (client *Client) HandleResponse() {
	io.Copy(os.Stdout, client.Conn)
	// 等价于
	// for {
	// 	buf := make([]byte, 4096)
	// 	client.Conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

func (client *Client) Rename() bool {
	fmt.Println(">>>>>>请输入新用户名<<<<<<")
	fmt.Scanln(&client.Name)

	sendMsg := fmt.Sprintf("rename|%s\n", client.Name)
	_, err := client.Conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn write: ", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	fmt.Println(">>>>>>请输入聊天内容,exit退出<<<<<<")
	var msg string
	fmt.Scanln(&msg)
	for msg != "exit" {
		if len(msg) > 0 {
			_, err := client.Conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Println("client conn write: ", err)
				break
			}
		}
		fmt.Println(">>>>>>请输入聊天内容,exit退出<<<<<<")
		msg = ""
		fmt.Scanln(&msg)
	}
}

func (client *Client) PrivateChat() {
	var callee string
	var msg string
	fmt.Println(">>>>>>请选择通信对象用户名,exit退出<<<<<<")
	client.GetOnlineUsers()
	fmt.Scanln(&callee)
	for callee != "exit" {
		fmt.Println(">>>>>>请输入聊天内容,exit退出<<<<<<")
		fmt.Scanln(&msg)
		for msg != "exit" {
			if len(msg) > 0 {
				_, err := client.Conn.Write([]byte(fmt.Sprintf("to|%s|%s\n", callee, msg)))
				if err != nil {
					fmt.Println("client conn write: ", err)
					break
				}
			}
			fmt.Println(">>>>>>请输入聊天内容,exit退出<<<<<<")
			msg = ""
			fmt.Scanln(&msg)
		}
		fmt.Println(">>>>>>请选择通信对象用户名,exit退出<<<<<<")
		client.GetOnlineUsers()
		fmt.Scanln(&callee)
	}

}

func (client *Client) GetOnlineUsers() bool {
	fmt.Println("<<<<<<当前在线人员>>>>>>")
	_, err := client.Conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("client conn write: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for !client.menu() {
		}
		switch client.flag {
		case 1:
			client.PublicChat()
		case 2:
			client.PrivateChat()
		case 3:
			client.Rename()
		case 4:
			client.GetOnlineUsers()
		default:
			fmt.Println("ByeBye…")
		}
	}
}

func main() {
	flag.Parse()
	client := NewClient(serverHost, serverPort)
	if client == nil {
		fmt.Println(">>>>>>服务器连接失败……")
		return
	}

	fmt.Println(">>>>>>服务器连接成功")

	go client.HandleResponse()

	// 阻塞
	// select{}
	client.Run()
}
