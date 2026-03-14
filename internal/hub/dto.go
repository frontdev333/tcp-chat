package hub

import "frontdev333/tcp-chat/internal"

type Request interface {
	execute(map[*internal.Client]bool)
}

type ActiveClientsIDsRequest struct {
	Response chan ActiveClientsIDsResult
}

type ActiveClientsIDsResult []string

type ClientsCountRequest struct {
	Response chan ClientsCountResult
}

type ClientsCountResult int

type ActiveClientsRequest struct {
	Response chan ActiveClientsResult
}

type ActiveClientsResult []*internal.Client

func (r *ActiveClientsIDsRequest) execute(clients map[*internal.Client]bool) {
	res := make([]string, 0, len(clients))

	for c := range clients {
		res = append(res, c.ID)
	}

	r.Response <- res
}

func (r *ClientsCountRequest) execute(clients map[*internal.Client]bool) {
	r.Response <- ClientsCountResult(len(clients))
}

func (a *ActiveClientsRequest) execute(clients map[*internal.Client]bool) {
	res := make([]*internal.Client, len(clients))

	i := 0
	for c, _ := range clients {
		res[i] = c
		i++
	}

	a.Response <- res
}
