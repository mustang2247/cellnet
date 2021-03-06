package socket

import (
	"net"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/gamedef"
)

type socketAcceptor struct {
	*peerBase
	*sessionMgr

	listener net.Listener

	running bool
}

func (self *socketAcceptor) Start(address string) cellnet.Peer {

	ln, err := net.Listen("tcp", address)

	self.listener = ln

	if err != nil {

		log.Errorf("#listen failed(%s) %v", self.name, err.Error())
		return self
	}

	self.running = true

	log.Infof("#listen(%s) %s ", self.name, address)

	// 接受线程
	go func() {
		for self.running {
			conn, err := ln.Accept()

			if err != nil {
				log.Errorf("#accept failed(%s) %v", self.name, err.Error())
				self.Post(self, newSessionEvent(Event_SessionAcceptFailed, nil, &gamedef.SessionAcceptFailed{Reason: err.Error()}))
				break
			}

			// 处理连接进入独立线程, 防止accept无法响应
			go func() {

				ses := newSession(NewPacketStream(conn), self, self)

				// 添加到管理器
				self.sessionMgr.Add(ses)

				// 断开后从管理器移除
				ses.OnClose = func() {
					self.sessionMgr.Remove(ses)
				}

				log.Infof("#accepted(%s) sid: %d", self.name, ses.ID())

				// 通知逻辑
				self.Post(self, NewSessionEvent(Event_SessionAccepted, ses, nil))
			}()

		}

	}()

	return self
}

func (self *socketAcceptor) Stop() {

	if !self.running {
		return
	}

	self.running = false

	self.listener.Close()
}

func NewAcceptor(evq cellnet.EventQueue) cellnet.Peer {

	self := &socketAcceptor{
		sessionMgr: newSessionManager(),
		peerBase:   newPeerBase(evq),
	}

	return self
}
