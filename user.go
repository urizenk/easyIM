package main

import (
	"math/rand"
	"strings"

	"net"
)

type User struct {
	Name string
	Addr string
	c chan string
	conn net.Conn
	server *Server
}


func NewUser(conn net.Conn, server *Server) *User{
	userAddr := conn.RemoteAddr().String()

	userName := randowName(5)

	user := &User{
		Name: userName,
		Addr: userAddr,
		c: make(chan string),
		conn: conn,
		server: server,

	}

	go user.listenMessage()
	return user


}

//从管道中读取消息发送给服务器
func (this *User)  listenMessage() {
	for{
		msg := <-this.c
		this.conn.Write([]byte(msg + "/n"))
	}
}
//用户上线
func (this *User) online()  {
	//保证用户表添加的原子性
	this.server.mapLock.Lock()
	this.server.onlineMap[this.Name]=this
	this.server.mapLock.Unlock()

	//广播上线消息
	this.server.broadCast(this,"上线了")
}

//给服务端用户对应客户端发消息
func (this *User) sendMessage(msg string)  {
	this.conn.Write([]byte(msg))
}

//用户处理消息
func (this *User) doMessage(msg string)  {
	//查询在线用户
	if msg == "who"{
		this.server.mapLock.Lock()
		for _,cli:= range this.server.onlineMap{
			onlineMsg := "[ " + cli.Addr + cli.Name + "在线" + " ]"
			cli.sendMessage(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|"{
		newName := strings.Split(msg, "|")[1]

		_, ok := this.server.onlineMap[newName]
		if ok{
			this.sendMessage("当前用户名已经被使用\n")
		}else {
			this.server.mapLock.Lock()
			delete(this.server.onlineMap,this.Name)
			this.server.onlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.sendMessage("用户名已经更改为"+ newName + "\n")
		}
	} else {
		this.server.broadCast(this,msg)
	}

}
//用户下线
func (this *User) offline()  {
	this.server.mapLock.Lock()
	delete(this.server.onlineMap,this.Name)
	this.server.mapLock.Unlock()

	//广播上线消息
	this.server.broadCast(this,"下线了")
}


//生成随机的用户名
func randowName(num int) string {
	charList := [26]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	var userName string
	for i := 0; i < num; i++ {
		userName += charList[rand.Intn(26)]
	}
	return userName
}
