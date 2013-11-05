package main

import (
	"log"
)

import (
	"agent/gsdb"
	"agent/hub_client"
	"db/forward_tbl"
	. "types"
)

//----------------------------------------------- connection close cleanup work
func close_work(sess *Session) {
	if sess.LoggedIn {
		hub_client.Logout(sess.User.Id)
		gsdb.UnregisterOnline(sess.User.Id)
		close(sess.MQ)

		// 未处理的IPC数据，重新放入db
		for ipcobject := range sess.MQ {
			forward_tbl.Push(&ipcobject)
			log.Println("re-push ipcobject back to db")
		}

		// 持久化逻辑#3: 离线时，刷入数据库
		_flush(sess)
	}
}
