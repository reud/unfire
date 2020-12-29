package service

import (
	"fmt"
	"time"
	"unfire/infrastructure/persistence"
)

type BatchService interface {
	Run()
}

type batchService struct {
	interval time.Duration
	ds       persistence.Datastore
}

func NewBatchService(interval time.Duration, ds persistence.Datastore) BatchService {
	return &batchService{
		interval: interval,
		ds:       ds,
	}
}

func (bs *batchService) Run() {
	ticker := time.NewTicker(bs.interval)
	go func() {
		for t := range ticker.C {
			fmt.Printf("batch started: %+v", t)

		}
	}()
}

func (bs *batchService) runTask() {

}
