/*
# @Time : 2020-06-02 12:01
# @Author : smallForest
# @SoftWare : GoLand
*/
package impl

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn    *websocket.Conn // 连接
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan []byte

	mutex    sync.Mutex
	isClosed bool // 保证线程安全 需要加锁mutex
}

// 初始化连接
func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000), // 缓冲1000
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan []byte, 1),
	}
	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()
	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed")

	}
	return
}
func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}
func (conn *Connection) Close() {
	// 线程安全的close 可以重入(多次调用)
	_ := conn.wsConn.Close()

	// 关闭channel 只可关闭一次
	conn.mutex.Lock() // 加锁
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()

}

// 内部实现 读长长连接的消息当道队列
func (conn *Connection) readLoop() {
	// 不停的读消息
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		// 阻塞在这里，等到inChan有空余位置
		conn.inChan <- data
		select { // 用于优雅退出
		case conn.inChan <- data:
		case <-conn.closeChan:
			// 当closeChan被关闭的时候
			goto ERR

		}

	}
ERR:
	conn.Close()
}
func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR

		}
		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
