package gcluster

import "github.com/csby/gwsf/gtype"

func (s *Controller) Write(message *gtype.SocketMessage, token *gtype.Token) {
	if message == nil {
		return
	}

	items := s.instances
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}

		out := item.Out
		if out == nil {
			continue
		}
		if !out.Connected() {
			continue
		}

		ch := out.Channel
		if ch == nil {
			continue
		}

		ch.Write(message)
	}
}

func (s *Controller) WriteMessage(message *gtype.SocketMessage, tokenId string) bool {
	if message == nil {
		return false
	}

	count := 0
	items := s.instances
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}
		if item.Index == s.index {
			continue
		}

		out := item.Out
		if out == nil {
			continue
		}
		if !out.Connected() {
			continue
		}

		ch := out.Channel
		if ch == nil {
			continue
		}

		t := ch.Token()
		if t != nil {
			if t.ID == tokenId {
				ch.Write(message)
				count++
			}
		}
	}

	return count > 0
}

func (s *Controller) WriteMsg(message *gtype.SocketMessage, filter gtype.SocketFilter) int {
	if message == nil {
		return 0
	}

	count := 0
	items := s.instances
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}
		if item.Index == s.index {
			continue
		}

		out := item.Out
		if out == nil {
			continue
		}
		if !out.Connected() {
			continue
		}

		ch := out.Channel
		if ch == nil {
			continue
		}

		if filter == nil {
			ch.Write(message)
			count++
		} else {
			if !filter.Ignored(ch.Token()) {
				ch.Write(message)
				count++
			}
		}

	}

	return count
}
