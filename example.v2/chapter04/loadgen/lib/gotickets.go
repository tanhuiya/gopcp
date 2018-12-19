package lib

import (
	"fmt"
	"github.com/pkg/errors"
)

type GoTickets interface {
	// 拿走一张
	Take()
	// 归还一张
	Return()
	// 票池是否激活
	Active() bool
	// 票池总数
	Total() uint32
	// 票池剩余
	Remainder() uint32
}

type myGoTickets struct {
	total  		uint32
	ticketCh 	chan struct{}
	active 		bool
}

func NewGoTicket(total uint32) (GoTickets, error) {
	gt := myGoTickets{}
	if !gt.init(total) {
		return nil, errors.New(fmt.Sprintf("The goroutine ticket pool can not be initialized, total %d\n", total))
	}
	return &gt, nil
}

func (gt *myGoTickets)init(total uint32) bool {
	if gt.active {
		return false
	}
	if total ==0 {
		return false
	}
	ch := make(chan struct{}, total)
	n := int(total)
	for i := 0; i < n ; i++ {
		ch <- struct{}{}
	}
	gt.ticketCh = ch
	gt.total = total
	gt.active = true
	return true
}

// 拿走一张
func (gt *myGoTickets)Take() {
	<-gt.ticketCh
}
// 归还一张
func (gt *myGoTickets)Return() {
	gt.ticketCh <- struct{}{}
}
// 票池是否激活
func (gt *myGoTickets)Active() bool {
	return gt.active
}
// 票池总数
func (gt *myGoTickets)Total() uint32 {
	return gt.total
}
// 票池剩余
func (gt *myGoTickets)Remainder() uint32 {
	return uint32(len(gt.ticketCh))
}