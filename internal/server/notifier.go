package server

import "internet-monitor/internal/types"

type Notifier struct{ s *Server }

func NewNotifier(s *Server) *Notifier { return &Notifier{s: s} }

func (n *Notifier) OnTick(result types.CheckResult, status types.Status) {
	n.s.UpdateTick(result, status)
}

func (n *Notifier) OnEvent(event types.Event) { n.s.AddEvent(event) }
