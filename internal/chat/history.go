package chat

import (
	"frontdev333/tcp-chat/internal/domain"
	"sync"
)

type History struct {
	lastMessages []*domain.ChatMessage
	head         uint
	tail         uint
	mtx          sync.Mutex
}

func NewHistory(size int) *History {
	return &History{
		lastMessages: make([]*domain.ChatMessage, size),
		head:         0,
		tail:         0,
	}
}

func (h *History) Add(msg *domain.ChatMessage) {
	capacity := uint(len(h.lastMessages))
	h.mtx.Lock()
	defer h.mtx.Unlock()
	index := h.head & (capacity - 1)
	h.head++
	h.lastMessages[index] = msg
}

func (h *History) Read() *domain.ChatMessage {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	capacity := uint(len(h.lastMessages))
	index := h.tail & (capacity - 1)
	h.tail++
	res := h.lastMessages[index]
	return res
}

func (h *History) GetLastMessages() []*domain.ChatMessage {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	capacity := uint(len(h.lastMessages))
	start := uint(0)

	if h.head > capacity {
		start = h.head - capacity
	}

	count := h.head - start
	res := make([]*domain.ChatMessage, 0, count)

	for i := start; i < h.head; i++ {
		index := i & (capacity - 1)
		res = append(res, h.lastMessages[index])
	}

	return res
}
