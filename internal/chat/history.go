package chat

import "sync"

type History struct {
	lastMessages []*ChatMessage
	head         uint
	tail         uint
	mtx          sync.Mutex
}

func NewHistory(size int) *History {
	return &History{
		lastMessages: make([]*ChatMessage, size),
		head:         0,
		tail:         0,
	}
}

func (h *History) Add(msg *ChatMessage) {
	capacity := uint(len(h.lastMessages))
	h.mtx.Lock()
	defer h.mtx.Unlock()
	index := h.head & (capacity - 1)
	h.head++
	h.lastMessages[index] = msg
}

func (h *History) Read() *ChatMessage {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	capacity := uint(len(h.lastMessages))
	index := h.tail & (capacity - 1)
	h.tail++
	res := h.lastMessages[index]
	return res
}

func (h *History) GetLastMessages() []*ChatMessage {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	capacity := uint(len(h.lastMessages))
	start := uint(0)

	if h.head > capacity {
		start = h.head - capacity
	}

	count := h.head - start
	res := make([]*ChatMessage, 0, count)

	for i := start; i < h.head; i++ {
		index := i & (capacity - 1)
		res = append(res, h.lastMessages[index])
	}

	return res
}
