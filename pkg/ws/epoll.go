package ws

import (
	log "github.com/sirupsen/logrus"
	"github.com/whyxn/go-chat-backend/pkg/core"
	"golang.org/x/sys/unix"
	"net"
	"reflect"
	"sync"
	"syscall"
)

/*
**** This file might show you some unresolved types or functions. These are unix based types. Will only resolve them in a Linux environment.
 */

type epoll struct {
	fd      int
	clients map[int]core.Client
	lock    *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:      fd,
		lock:    &sync.RWMutex{},
		clients: make(map[int]core.Client),
	}, nil
}

func (e *epoll) Add(client core.Client) error {
	// Extract file descriptor associated with the connection
	fd := websocketFD(client.Connection)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.clients[fd] = client
	if len(e.clients)%100 == 0 {
		log.Info("Total number of connections: %v", len(e.clients))
	}
	return nil
}

func (e *epoll) Remove(client core.Client) error {
	fd := websocketFD(client.Connection)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.clients, fd)

	core.GetChatEngine().GetGoChat().RemoveClient(client)
	log.Info("Total clients", len(core.GetChatEngine().GetGoChat().ClientsMap))

	if len(e.clients)%100 == 0 {
		log.Info("Total number of connections: %v", len(e.clients))
	}
	return nil
}

func (e *epoll) Wait() ([]core.Client, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		return nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var clients []core.Client
	for i := 0; i < n; i++ {
		client := e.clients[int(events[i].Fd)]
		clients = append(clients, client)
	}
	return clients, nil
}

func websocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
