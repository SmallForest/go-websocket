/*
# @Time : 2020-06-02 11:21
# @Author : smallForest
# @SoftWare : GoLand
*/
package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go-websocket/impl"
	"net/http"
	"time"
)

var (
	updrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn *websocket.Conn
		err    error
		data   []byte
		conn   *impl.Connection
	)
	// 1 收到http请求转换成ws
	if wsConn, err = updrader.Upgrade(w, r, nil); err != nil {
		return
	}
	// 将原始的webcoket进行封装
	if conn, err = impl.InitConnection(wsConn); err != nil {
		goto ERR
	}
	// 心跳检测
	go func() {
		var (
			err error
		)
		for {
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}

	}()
	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		if err = conn.WriteMessage(data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
func main() {
	// http://localhost:7777/ws
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("ws://localhost:7777/ws")
	http.ListenAndServe("0.0.0.0:7777", nil)
}
