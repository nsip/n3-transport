// worker.go

package n3dispatcher

import (
	"context"
	"fmt"
	"log"

	"../n3crypto"
	lbproto "github.com/liftbridge-io/go-liftbridge/liftbridge-grpc"
	"github.com/nsip/n3-messages/messages"
	"github.com/nsip/n3-messages/messages/pb"
)

type workerConfig struct {
	ctx        context.Context
	user       string
	scope      string
	handlerTag string
	disp       *Dispatcher
}

func startWorker(cfg *workerConfig) error {

	// handler
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("worker received error from liftbridge: ", err)
			return
		}

		n3msg, err := messages.DecodeN3Message(msg.Value)
		if err != nil {
			log.Println("worker unable to decode message: ", err)
			return
		}

		// check msg routing to see if we're interested
		msgScope := fmt.Sprintf("%s.%s", n3msg.NameSpace, n3msg.CtxName)
		if msgScope != cfg.scope {
			// log.Println("worker received msg for", msgScope, " ignoring.")
			return
		}
		// log.Println("worker received msg for", msgScope, " will process.")

		// if we are, decrypt the payload
		// and re-crypt for the target user
		tuple, err := n3crypto.DecryptTuple(n3msg.Payload, n3msg.NameSpace, cfg.disp.privKey)
		if err != nil {
			log.Println("worker unable to decrypt message tuple: ", err)
			return
		}
		recrypted, err := n3crypto.EncryptTuple(tuple, cfg.user, cfg.disp.privKey)
		if err != nil {
			log.Println("encryption error in worker: ", err)
			return
		}

		// build new message
		newMsg := &pb.N3Message{
			Payload:   recrypted,
			SndId:     n3msg.SndId,
			NameSpace: n3msg.NameSpace,
			CtxName:   n3msg.CtxName,
			DispId:    cfg.disp.pubKey,
		}
		msgBytes, err := messages.EncodeN3Message(newMsg)

		// send
		err = cfg.disp.natsConn.Publish(cfg.user, msgBytes)
		if err != nil {
			log.Println("worker failed to publish message: ", err)
			return
		}
		// log.Println("worker sent message to: ", cfg.user)

	}

	go func() {

		// subscribe
		streamName := fmt.Sprintf("%s-stream", cfg.disp.pubKey)
		err := cfg.disp.lbConn.Subscribe(cfg.ctx, cfg.disp.pubKey, streamName, handler)
		if err != nil {
			log.Println("error subscribing worker: ", err)
			cfg.disp.removeHandlerContext(cfg.handlerTag)
		}
		<-cfg.ctx.Done()
		log.Println("worker closing for: ", cfg.user, cfg.scope)

		// TODO: persist last stream position to prevent data re-sends

	}()

	return nil
}
