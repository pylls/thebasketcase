package main

import (
	"github.com/pylls/thebasketcase/model"
	"golang.org/x/net/context"
)

type server struct{}

func (s *server) Work(c context.Context,
	in *model.Report) (out *model.Browse, err error) {
	lock.Lock() // ugly, but good enough
	defer lock.Unlock()

	// keep tabs on number of workers
	_, exists := workers[in.WorkerID]
	if !exists {
		workers[in.WorkerID] = in.WorkerID
	}

	if in.Browse.ID != "" && in.Browse.BatchID == batchID && // work in batch?
		len(in.Pcap) >= *minDataLen { // and did we get enough data to be happy?
		err = store(in) // OK, store completed work
		if err != nil {
			return nil, err
		}
	}

	if len(work) == 0 { // no work right now
		return &model.Browse{
			ID:      "", // empty ID -> sleep for timeout seconds
			Timeout: int64(*timeout),
		}, nil
	}

	i := <-work
	return &model.Browse{
		ID:      i.ID,
		BatchID: batchID,
		URL:     i.URL,
		Torrc:   torrc,
		Log:     *getTorLog,
		Timeout: int64(*timeout),
	}, nil
}
