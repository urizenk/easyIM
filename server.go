package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int
	onlineMap map[string] *User
	mapLock sync.RWMutex
	message chan string
}
//构造一个新的server对象
func newServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		onlineMap: make(map[string]*User),
		message: make(chan string),
	}

	return server
}

func (this *Server ) listenMessageChan()  {
	for{
		msg := <-this.message
		this.mapLock.Lock()
		//从用户表中取出客户端，将
		for _, cli := range this.onlineMap{
			cli.c <- msg
		}
		this.mapLock.Unlock()
	}

}

//处理连接业务
func (this *Server) Handler(conn net.Conn) {

	//接收用户对象
	user := NewUser(conn,this)

	fmt.Println("链接建立成功")

	//用户上线
	user.online()

	isAlive := make(chan bool)

	//读取用户消息并广播

	go func() {
		buf := make([]byte, 4096)

		for{
			n,err := conn.Read(buf)
			if n == 0{
				user.offline()
				return
			}
			if err != nil && err != io.EOF{
				fmt.Println("Conn err ",err )
				return
			}

			msg := string(buf)


			user.doMessage(msg)
			//每当发送消息时像存活状态管道发送一个消息
			isAlive <- true
		}
	}()

	for{
		select {
			case <- isAlive:
				//此时什么都不做，会自动执行下面的case,重置定时器
			case <- time.After(time.Second * 20):
				user.sendMessage("你由于长时间没有活动，已经被踢出服务器")
				//回收资源
				close(user.c)
				conn.Close()
				return
		}
	}
}
//广播用户上线消息
func (this *Server) broadCast(user *User, message string)  {
	sendMsg := "[ " + user.Addr + user.Name + message + " ]"
	this.message<-sendMsg
}

//服务器开启
func (this *Server) Start() {

	//listen socket
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listener err", err)
	}

	//启动监听message管道的go程

	go this.listenMessageChan()

	//close socket
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("close err", err)
		}
	}(listener)

	for {
		//accept event
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.accept error", err)
			continue
		}

		go this.Handler(conn)

	}

}
