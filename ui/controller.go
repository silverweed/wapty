//ui is a general high level representation of all the uis connected to the current
//instance of Wapty. Use this from other packages to read user input and write
//output
package ui

import (
	"log"
	"sync"

	"github.com/empijei/wapty/ui/apis"
)

//String used to detect the main action within an ui.Args
const ACTION = "action"

type SubsChannel map[int64]Subscription

//TODO Use a map[chan apis.Command]struct{} if nothing else than the channel is used
type UI struct {
	id      int
	channel chan apis.Command
}

func (u *UI) Channel() <-chan apis.Command {
	return u.channel
}

var subScriptions map[string]SubsChannel
var subsMutex sync.RWMutex
var subsCounter int64
var iChan chan apis.Command
var oChans uis

type uis struct {
	sync.RWMutex
	curID int
	list  map[int]*UI
}

func init() {
	subScriptions = make(map[string]SubsChannel)
	iChan = make(chan apis.Command)
	oChans.list = make(map[int]*UI)
}

type Subscription struct {
	id          int64
	channel     string
	dataCh      chan apis.Command
	DataChannel <-chan apis.Command
}

func Subscribe(channel string) *Subscription {
	subsMutex.Lock()
	subsCounter++
	//Unless you are sure the out channel will be constantly read, it is strongly
	//suggested to create a buffered channel
	pipe := make(chan apis.Command, 50) //TODO this is arbitrary, give a meaning to this number
	out := Subscription{id: subsCounter, dataCh: pipe, channel: channel}
	if subScriptions[channel] == nil {
		subScriptions[channel] = make(map[int64]Subscription)
	}
	subScriptions[channel][subsCounter] = out
	out.DataChannel = pipe
	subsMutex.Unlock()
	return &out
}

//Sends the command and sets the channel with the value set in the subscription
func (s *Subscription) Send(c apis.Command) {
	c.Channel = s.channel
	send(c)
}

//TODO delete this? Dangerous and never used
func UnSubscribe(s *Subscription) {
	subsMutex.RLock()
	defer subsMutex.RUnlock()
	for _, channelSubs := range subScriptions {
		sub, ok := channelSubs[s.id]
		if ok {
			subsMutex.Lock()
			close(sub.dataCh)
			delete(channelSubs, s.id)
			subsMutex.Unlock()
			return
		}
	}
	log.Println("Subscription not found")
}

func send(c apis.Command) {
	oChans.RLock()
	for _, oChan := range oChans.list {
		oChan.channel <- c
	}
	oChans.RUnlock()
}

//This should be a server's method
func Receive(c apis.Command) {
	iChan <- c
}

//This function is used by uis servers to read all the messages from wapty and send them to clients.
func Connect() *UI {
	oChans.Lock()
	defer oChans.Unlock()
	toRet := &UI{channel: make(chan apis.Command), id: oChans.curID}
	oChans.list[oChans.curID] = toRet
	oChans.curID++
	return toRet
}

func Disconnect(u *UI) {
	oChans.Lock()
	defer oChans.Unlock()
	delete(oChans.list, u.id)
}

func ControllerMainLoop() {
	for cmd := range iChan {
		subsMutex.RLock()
		for _, out := range subScriptions[cmd.Channel] {
			out.dataCh <- cmd
		}
		subsMutex.RUnlock()
	}
}
