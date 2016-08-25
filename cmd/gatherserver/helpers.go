package main

import (
	"io/ioutil"
	"path"

	"github.com/pylls/thebasketcase/model"
)

func store(in *model.Browse) (err error) {
	if len(in.Data) > 0 {
		err = ioutil.WriteFile(outputFileName(in.ID), in.Data, 0666)
		if err == nil {
			done++
		}
	}
	return
}

func outputFileName(id string) string {
	return path.Join(*datadir, path.Clean(id)+*outputSuffix)
}
