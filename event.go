package fmp

import "github.com/panjf2000/gnet"

type IEventHandler interface {
	OnOpened(c gnet.Conn)
	OnClosed(c gnet.Conn)
	OnMn(mn string, connect bool)
	React(c gnet.Conn, msg *Msg)
}

type EventHandler struct{}

func (h *EventHandler) OnOpened(c gnet.Conn) {}

func (h *EventHandler) OnClosed(c gnet.Conn) {}

func (h *EventHandler) OnMn(mn string, connect bool) {}

func (h *EventHandler) React(c gnet.Conn, msg *Msg) {}
