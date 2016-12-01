package client

import (
	"fmt"
	"strings"
	"time"
  "github.com/nu7hatch/gouuid"

	"github.com/soundtrackyourbrand/pusher/hub"
)

type Client struct {
	idIncr   int
	id       string
	outgoing hub.OutgoingMessage
	incoming hub.IncomingMessage
	closed   chan bool
	closed2   chan bool
  receivers map[string](chan hub.Message)
  other chan hub.Message
}

func (self *Client) getNextId() string {
  channel, _ := uuid.NewV4()
	return channel.String()
}

func (self *Client) Connect(origin, location string) {
	self.outgoing, self.incoming = hub.Connect("", origin, location)
	self.closed = make(chan bool)
	self.closed2 = make(chan bool)
	self.other = make(chan hub.Message)
  self.receivers = make(map[string](chan hub.Message))
  welcome := hub.Message{}

  // Incomming loop
  go func() {
		for {
			select {
      case  msg, _ := <- self.incoming:
        if msg.Type == hub.TypeWelcome {
	        welcome = msg
	        self.id = welcome.Welcome.Id
        } else if msg.Type == hub.TypeHeartbeat {

        } else {
          r, found := self.receivers[msg.Id]
          if found {
            r <- msg
          } else {
            self.other <- msg
          }
        }
			case _, ok := <-self.closed:
				if !ok {
					break
				}
     }
   }
  }()

  // Outgoing loop
	go func() {
		defer func() {
			if e := recover(); e != nil {
				if !strings.Contains(fmt.Sprint(e), "closed channel") {
					panic(e)
				}
			}
		}()
		for {
			select {
			case self.outgoing <- hub.Message{Type: hub.TypeHeartbeat}:
			case _, ok := <-self.closed:
				if !ok {
					break
				}
			default:
				break
			}
      if (welcome.Welcome != nil) {
			  time.Sleep(time.Millisecond * welcome.Welcome.Heartbeat)
      } else {
			  time.Sleep(time.Millisecond * 1000)
      }
		}
	}()
}
func (self *Client) Close() {
	close(self.closed)
	close(self.closed2)
	close(self.outgoing)
}

func (self *Client) SendAndReceive(send hub.Message) (err error, recv hub.Message) {
  send.Id = "foo" + self.getNextId()
  self.receivers[send.Id] = make(chan hub.Message)
	self.outgoing <- send
  recv = <- self.receivers[send.Id]
  delete(self.receivers, send.Id)
  return
}

func (self *Client) Authorize(uri, token string) (err error) {
  self.SendAndReceive(hub.Message{Type: hub.TypeAuthorize, URI: uri, Token: token, Write: true})
	return
}

func (self *Client) Subscribe(uri string) (err error) {
	self.SendAndReceive(hub.Message{Type: hub.TypeSubscribe, URI: uri})
	return
}

func (self *Client) Unsubscribe(uri string) (err error) {
	self.SendAndReceive(hub.Message{Type: hub.TypeUnsubscribe, URI: uri})
	return
}

// {"Type":"Message","URI":"/devices/RGV2aWNlLCwxbWptcjU3aDVhOC8./rpcclient/rpcserver/webserver","Data":[{"uri":"/logfilter.html","channel":"f4309837-4087-1000-0bea-b1a94acfb63d"}],"Id":"H1e77lWU21pBddNhPy4c-Q==:4"}
func (self *Client) Send(uri string, data interface{}) (err error) {
	self.SendAndReceive(hub.Message{Type: hub.TypeMessage, Data: data, URI: uri})
	return
}

func (self *Client) Next() hub.Message {
	return <- self.other
}


