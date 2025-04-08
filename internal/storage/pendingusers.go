package storage

import (
	"sync"
	"time"
)

// UserTimer ожидающие пользователи
type UserTimer struct {
	Timer  *time.Timer
	ChatID int64
	UserID int64
}

type PendingUser struct {
	values map[int64]*UserTimer
	mx     sync.RWMutex
}

func NewPendingUser() *PendingUser {
	instance := new(PendingUser)
	instance.values = make(map[int64]*UserTimer)
	return instance
}

func (p *PendingUser) Delete(userID int64) {
	p.mx.Lock()
	defer p.mx.Unlock()
	delete(p.values, userID)
}
func (p *PendingUser) Get(userID int64) (*UserTimer, bool) {
	p.mx.Lock()
	defer p.mx.Unlock()
	v, ok := p.values[userID]
	return v, ok
}

func (p *PendingUser) Set(userID int64, timer *UserTimer) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.values[userID] = timer
}
