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

	// store the data if:
	// the worker is reporting work (ID not empty),
	// the work is in the active batch,
	// we collected enough data, and
	// we're sure we haven't collected enough data already
	if in.Browse != nil && in.Browse.ID != "" && in.Browse.BatchID == batchID &&
		len(in.Pcap) >= *minDataLen &&
		(len(done) < *monitored**samples || len(done) < *unmonitored) {
		err = store(in)
		if err != nil {
			return nil, err
		}

		done[in.Browse.ID] = true
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
