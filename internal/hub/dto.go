package hub

import "frontdev333/tcp-chat/internal"

type Request interface {
	execute(map[*internal.Client]bool)
}

type ActiveClientsResult []string

type ActiveClientsRequest struct {
	Response chan ActiveClientsResult
}

type ClientsCountRequest struct {
	Response chan ClientsCountResult
}

type ClientsCountResult int

func (r *ActiveClientsRequest) execute(clients map[*internal.Client]bool) {
	res := make([]string, 0, len(clients))

	for c := range clients {
		res = append(res, c.ID)
	}

	r.Response <- res
}

func (r *ClientsCountRequest) execute(clients map[*internal.Client]bool) {
	r.Response <- ClientsCountResult(len(clients))
}
