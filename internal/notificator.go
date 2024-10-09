package internal

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrNotifiyChannelShutDown error = errors.New("Notify Channel Maybe shutdown!")
	ErrNotifyReceiverDisable  error = errors.New("Notify receiver disable")
)

type Notificator struct {
	channels map[uuid.UUID]chan interface{}
	notifyMu *sync.RWMutex
}

func NewNotificator() *Notificator {
	return &Notificator{
		channels: make(map[uuid.UUID]chan interface{}),
		notifyMu: &sync.RWMutex{},
	}
}

func (this *Notificator) Create(bufSize int) (<-chan interface{}, uuid.UUID) {
	id := uuid.New()
	c := make(chan interface{}, bufSize)
	this.notifyMu.Lock()
	this.channels[id] = c
	this.notifyMu.Unlock()
	return c, id
}

func (this *Notificator) Remove(id uuid.UUID) error {
	this.notifyMu.Lock()
	defer this.notifyMu.Unlock()
	if c, ok := this.channels[id]; ok {
		delete(this.channels, id)
		close(c)
		return nil
	}
	return ErrNotifiyChannelShutDown
}

func (this *Notificator) Notify(id uuid.UUID, v interface{}, blocking bool) error {
	this.notifyMu.RLock()
	defer this.notifyMu.RUnlock()

	if c, ok := this.channels[id]; ok {
		if blocking {
			c <- v
		} else {
			select {
			case c <- v:
			default:
				return ErrNotifyReceiverDisable
			}
		}
		return nil
	}
	return ErrNotifyReceiverDisable
}
