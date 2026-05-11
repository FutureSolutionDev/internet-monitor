package dashboard

import "internet-monitor/types"

// DashNotifier forwards monitoring events to the dashboard SSE stream.
type DashNotifier struct {
	s *Server
}

func NewNotifier(s *Server) *DashNotifier {
	return &DashNotifier{s: s}
}

func (n *DashNotifier) OnTick(result types.CheckResult, status types.Status) {
	n.s.UpdateTick(result, status)
}

func (n *DashNotifier) OnEvent(event types.Event) {
	n.s.AddEvent(event)
}
