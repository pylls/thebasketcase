package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/pylls/thebasketcase/model"
	"golang.org/x/net/context"
)

type server struct{}

func (s *server) Work(c context.Context,
	in *model.Req) (out *model.Browse, err error) {
	lock.Lock() // ugly, but good enough
	defer lock.Unlock()

	// keep tabs on number of workers
	_, exists := workers[in.WorkerID]
	if !exists {
		workers[in.WorkerID] = in.WorkerID
		fmt.Println("")
		log.Printf("worker reporting for work: %s\n", in.WorkerID)
	}

	if in.Browse.ID != "" { // completed work?
		if len(in.Browse.Data) >= *minDataLen {
			err = store(in.Browse)
			if err != nil {
				work <- item{ // put work back in
					ID:  in.Browse.ID,
					URL: in.Browse.URL,
				}
				return
			}

		} else { // too little data, put back
			// put back work, toggling "www.": this helps us to reach
			// sites that for some reason only is accessible with/without www,
			// like www.googleusercontent.com listed as googleusercontent.com
			// in Alexa top-1m.
			url := in.Browse.URL
			if strings.Contains(url, "www.") {
				url = strings.Replace(url, "www.", "", 1)
			} else {
				url = strings.Replace(url, "://", "://www.", 1)
			}

			work <- item{ // this overwrites what we have
				ID:  in.Browse.ID,
				URL: url,
			}
		}
	}

	if len(work) == 0 {
		// no work right now
		return &model.Browse{
			ID:      "",
			Timeout: int64(*timeout),
		}, nil
	}

	i := <-work
	return &model.Browse{
		ID:      i.ID,
		URL:     i.URL,
		Timeout: int64(*timeout),
	}, nil

}
