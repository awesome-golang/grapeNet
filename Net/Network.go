// 网络层可以使用多种协议
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/10

package grapeNet

import (
	"net"

	cm "github.com/koangel/grapeNet/ConnManager"
	logger "github.com/koangel/grapeNet/Logger"
)

type NotifyApi interface {
	// 创建用户DATA
	CreateUserData() interface{}

	// 通知连接
	OnAccept(sid string, conn *TcpConn)
	OnHandler(sid string, conn *TcpConn)
	OnClose(sid string, conn *TcpConn)
	OnConnected(sid string, conn *TcpConn)

	MainProc() // 简易主处理函数
}

type TCPNetwork struct {
	listener net.Listener

	DialCM   *cm.ConnManager
	AcceptCM *cm.ConnManager

	notifyBind NotifyApi // 在指定的行为中通知该服务端
}

/////////////////////////////
// 创建网络服务器
func NewTcpServer(addr string, papi NotifyApi) (tcp *TCPNetwork, err error) {
	tcp = &TCPNetwork{
		listener:   nil,
		notifyBind: papi,
		DialCM:     cm.NewCM(),
		AcceptCM:   cm.NewCM(),
	}

	err = tcp.listen(addr)
	if err != nil {
		tcp = nil
	}
	return
}

////////////////////////////
// 成员函数
func (c *TCPNetwork) BindNotify(papi NotifyApi) {
	c.notifyBind = papi
}

func (c *TCPNetwork) Dial(addr string, UserData interface{}) (conn *TcpConn, err error) {
	logger.INFO("Dial To :%v", addr)
	conn, err = NewDial(c, addr, UserData)
	if err != nil {
		logger.ERROR("Dial Faild:%v", err)
		return
	}

	c.notifyBind.OnConnected(conn.SessionId, conn)
	c.DialCM.Register <- conn // 注册账户
	return
}

func (c *TCPNetwork) listen(bindAddr string) error {
	if c.listener != nil {
		return nil
	}

	lis, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return err
	}

	c.listener = lis
	go c.onAccept()
	return nil
}

/// 连接池的处理
func (c *TCPNetwork) onAccept() {
	defer func() {

	}()
	// 1000次错误 跳出去
	for failures := 0; failures < 1000; {
		conn, listenErr := c.listener.Accept()
		if listenErr != nil {
			failures++
			continue
		}

		logger.INFO("New Connection:%v，Accept.", conn.RemoteAddr())
		var client = NewConn(c, conn, c.notifyBind.CreateUserData())

		c.AcceptCM.Register <- client // 注册一个全局对象

		c.notifyBind.OnAccept(client.SessionId, client)

		client.startProc() // 启动线程
	}
}
