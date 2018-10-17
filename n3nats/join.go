// join.go

package n3nats

import (
	"log"
	"time"

	nats "github.com/nats-io/go-nats"
)

//
// Join pings the network to retrieve a
// dispatcher to work with
// nc - nats connection to usef for comms
// userid - b58 public key of the user wanting to connect
//
// returns string - b58 pub key of elected dispatcher
//
func Join(nc *nats.Conn, userid string) (string, error) {

	subj, payload := "join", []byte(userid)

	msg, err := nc.Request(subj, []byte(payload), 500*time.Millisecond)
	if err != nil {
		if nc.LastError() != nil {
			return "", err
		}
		return "", err
	}

	dispatcherid := string(msg.Data)

	log.Println("join ok")
	log.Println("received dispatcher id: ", dispatcherid)

	return dispatcherid, nil

}
