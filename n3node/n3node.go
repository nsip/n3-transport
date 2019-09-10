// n3node.go

package n3node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	liftbridge "github.com/liftbridge-io/go-liftbridge"
	lbproto "github.com/liftbridge-io/liftbridge-grpc/go"
	nats "github.com/nats-io/go-nats"
	"github.com/nsip/n3-messages/messages"
	"github.com/nsip/n3-messages/messages/pb"
	"github.com/nsip/n3-messages/n3grpc"
	"github.com/nsip/n3-transport/n3config"
	"github.com/nsip/n3-transport/n3crypto"
	"github.com/nsip/n3-transport/n3influx"
	"github.com/nsip/n3-transport/n3liftbridge"
	"github.com/nsip/n3-transport/n3nats"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// N3Node :
// main node tht handles read / write operations
// with the transport layer
//
type N3Node struct {
	natsConn         *nats.Conn
	lbConn           liftbridge.Client
	handlerContexts  map[string]func()
	dispatcherID     string
	mutex            sync.Mutex
	pubKey           string
	privKey          string
	approvedContexts map[string]bool
}

// NewNode :
func NewNode() (*N3Node, error) {

	err := checkConfig()
	if err != nil {
		return nil, errors.Wrap(err, "ensure you have run 'n3cli init'")
	}

	// get the b58 string of the public key
	b58pubkey := viper.GetString("pubkey")
	log.Println("this node id is: ", b58pubkey)
	b58privkey := viper.GetString("privkey")

	// create a nats client connection
	natsurl := viper.GetString("nats_addr")
	natsconn, err := nats.Connect(natsurl)
	if err != nil {
		return nil, errors.Wrap(err, "node cannot connect to Nats")
	}

	// create a liftbridge client connection
	lbaddr := viper.GetString("lb_addr")
	lbconn, err := liftbridge.Connect([]string{lbaddr})
	if err != nil {
		return nil, errors.Wrap(err, "node cannot connect to Liftbridge")
	}

	// initialise the handler contexts map
	handlercontexts := make(map[string]func())

	// initilaise the approved contexts map
	approvedContexts := make(map[string]bool)

	n3c := &N3Node{
		pubKey:           b58pubkey,
		privKey:          b58privkey,
		natsConn:         natsconn,
		lbConn:           lbconn,
		handlerContexts:  handlercontexts,
		approvedContexts: approvedContexts,
	}

	// ensure node read stream is available
	streamName := fmt.Sprintf("%s-stream", n3c.pubKey)
	err = n3liftbridge.CreateStream(lbaddr, n3c.pubKey, streamName)
	if err != nil {
		return nil, errors.Wrap(err, "node cannot connect to read stream")
	}

	//
	// listen for approvals and allow this node
	// to send to approved contexts
	//
	err = n3c.startApprovalHandler()
	if err != nil {
		n3c.Close()
		return nil, err
	}

	//
	// write handler listens for messages on the grpc server
	// for this client and sends to dispatcher
	//
	err = n3c.startWriteHandler()
	if err != nil {
		n3c.Close()
		return nil, err
	}

	//
	// reads all messages from this client's feed
	// and passes on to storage handlers
	//
	err = n3c.startReadHandler()
	if err != nil {
		return nil, err
	}

	return n3c, nil

}

//
// checks that a config file exists
//
func checkConfig() error {

	return n3config.ReadConfig()
}

//
// reads from the approvals stream and records the
// approval state for target streams this client can
// send messages to
//
func (n3c *N3Node) startApprovalHandler() error {

	//
	// logic invoked when an approval message is received
	//
	handler := func(msg *lbproto.Message, err error) {
		if err != nil {
			log.Println("node approval handler received error from liftbridge: ", err)
			return
		}

		// decode n3 message from transmission protobuf format
		// approvals are not point-to-point encrypted as they need
		// to be visible to all dispatchers and clients
		n3msg, err := messages.DecodeN3Message(msg.Value)
		if err != nil {
			log.Println("node cannot decode message proto: ", err)
			return
		}
		log.Println("node received approval from:", n3msg.SndId)

		// decode the internal tuple
		approvalTuple, err := messages.DecodeTuple(n3msg.Payload)
		if err != nil {
			log.Println("node cannot decode approval tuple proto: ", err)
			return
		}

		// record approval
		approvalScope := fmt.Sprintf("%s.%s.%s", approvalTuple.Subject, n3msg.NameSpace, approvalTuple.Object)
		log.Println("approval scope: ", approvalScope)
		n3c.mutex.Lock()
		defer n3c.mutex.Unlock()

		// handle the approval
		if approvalTuple.Predicate == "grant" {
			n3c.approvedContexts[approvalScope] = true
		}
		if approvalTuple.Predicate == "deny" {
			n3c.approvedContexts[approvalScope] = false
		}
	}

	// create context to control async subscription
	ctx, cancelFunc := context.WithCancel(context.Background())
	n3c.addHandlerContext("approvals", cancelFunc)

	// create the subscription and run until context is cancelled
	go func() {
		err := n3c.lbConn.Subscribe(ctx, "approvals-stream", handler, liftbridge.StartAtEarliestReceived())
		if err != nil {
			log.Println("node error subscribing approvals handler: ", err)
			n3c.removeHandlerContext("approvals")
		}
		<-ctx.Done()
	}()

	return nil
}

//
// listens for new messages from this client and sends to
// dispatcher
//
func (n3c *N3Node) startWriteHandler() error {

	// get an instance of direct db query
	dbClient, _ := n3influx.NewDBClient()

	// join network, acquire dispatcher
	dispatcherid, err := n3nats.Join(n3c.natsConn, n3c.pubKey)
	if err != nil {
		return errors.Wrap(err, "node closing: no dispatcher available.")
	}

	// Fetch Data to be Cached
	if dbClient.DbTblExists("databases", "tuples") {
		// Fetch Privacy Control Rule
		if dbClient.DbTblExists("measurements", "ctxid") && dbClient.DbTblExists("measurements", "privctrl") {
			pcObjCtxPathRW = mkPrivCtrl(dbClient)
		}
		// Fetch Meta Data
		mpObjIDVer = mkObjIDVerBuf(dbClient)
	}
	if mpObjIDVer == nil {
		mpObjIDVer = &sync.Map{}
	}

	// set up handler for inbound messages [SPREAD]
	mHandler := func(n3msg *pb.N3Message) {

		// force sender id to be this node
		n3msg.SndId = n3c.pubKey

		// unpack tuple payload
		tuple, err := messages.DecodeTuple(n3msg.Payload)
		if err != nil {
			log.Println("write handler cannot decode tuple: ", err)
			return
		}

		s, p, o, v, ctx := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version, n3msg.CtxName
		root := sSpl(p, DELIPath)[0]
		now := time.Now().Format("2006-01-02 15:04:05.000")

		// CHECK AUTHORISATION
		approvalScope := fmt.Sprintf("%s.%s.%s", n3msg.SndId, n3msg.NameSpace, ctx)
		if !n3c.approved(approvalScope) {
			log.Println("you are not authorised to send to context: ", n3msg.NameSpace, ctx)
			return
		}

		tupleQueue, ctxQueue := []*pb.SPOTuple{tuple}, []string{ctx} // *** Queue: 1.priv 2.ctxid || 1.data 2.meta ***

		// DONE: PRIVACY, STORE PRIVACY CONTROL FILE to <privctrl> & register it to <ctxid>
		if ctx == "privctrl" {
			if S(p).Contains("forcontext") { //      *** register to ctxid (privacy control file) ***
				vNext := dbClient.LatestVer("ctxid") + 1
				tupleQueue = append(tupleQueue, &pb.SPOTuple{Subject: root, Predicate: o, Object: s, Version: vNext})
				ctxQueue = append(ctxQueue, "ctxid")
				goto PUB
			}
			if !S(s).IsUUID() || S(o).HP("::") || S(o).HP("[]") { // *** ignore useless tuples ***
				return
			}
		}

		if ctx == "ctxid" { //                                       *** register / unregister to ctxid (cli) ***
			tupleQueue, ctxQueue = []*pb.SPOTuple{}, []string{}
			vNext := dbClient.LatestVer("ctxid") + 1
			tupleQueue = append(tupleQueue, &pb.SPOTuple{Subject: s, Predicate: p, Object: o, Version: vNext})
			ctxQueue = append(ctxQueue, "ctxid")
			goto PUB
		}

		// DONE: PRIVACY, CHECK PRIVACY RULES
		if ctx != "privctrl" && ctx != "ctxid" {
			if pcCtxPathCtrl, ok := pcObjCtxPathRW[root]; ok {
				if pcPathCtrl, ok1 := pcCtxPathCtrl[ctx]; ok1 {
					if ctrl, ok2 := pcPathCtrl[p]; ok2 {
						switch {
						case IArrEleIn(ctrl, Ss{"R/W", "W/R", "RW", "WR", "W"}):
						case S(ctrl).IsUUID():
						default:
							tupleQueue[0] = &pb.SPOTuple{Subject: s, Predicate: p, Object: NOWRITE, Version: v}
						}
					}
				}
			}
		}

		// DONE: ASSIGN TUPLE VERSION
		if p == MARKDead { //                                     *** Delete object ***
			if exist, alive := dbClient.Status(ctx, s); exist && alive {
				tupleV := &pb.SPOTuple{Subject: s, Predicate: p, Object: now, Version: -1}
				tupleS := &pb.SPOTuple{Subject: "::" + s, Predicate: p, Object: now, Version: -1}
				tupleA := &pb.SPOTuple{Subject: "[]" + s, Predicate: p, Object: now, Version: -1}
				ctxMeta := S(ctx).MkSuffix("-meta").V()
				tupleQueue = []*pb.SPOTuple{tupleV, tupleS, tupleA}
				ctxQueue = []string{ctxMeta, ctxMeta, ctxMeta}
			} else {
				return
			}
		} else {
			if metaTuple, metaCtx := assignVer(dbClient, tuple, ctx); metaTuple != nil { // *** Only at end of an whole object, can get metaTuple ***
				// *** Store prevID's low-high version map into meta context as <"id" "V/S/A" "low-high"> ***
				tupleQueue, ctxQueue = append(tupleQueue, metaTuple), append(ctxQueue, metaCtx)
			}
		}

	PUB:

		for i := 0; i < len(tupleQueue); i++ {

			// specify dispatcher
			// n3msg.DispId = dispatcherid

			// encrypt tuple payload for dispatcher
			encryptedTuple, err := n3crypto.EncryptTuple(tupleQueue[i], dispatcherid, n3c.privKey)
			if err != nil {
				log.Println("write handler unable to encrypt tuple: ", err)
				return
			}

			newMsg := &pb.N3Message{
				Payload:   encryptedTuple,
				SndId:     n3c.pubKey,
				NameSpace: n3msg.NameSpace,
				CtxName:   ctxQueue[i],
				DispId:    dispatcherid,
			}

			// encode & send
			msgBytes, err := messages.EncodeN3Message(newMsg)
			if err != nil {
				log.Println("write handler unable to encode message: ", err)
				return
			}

			// only for testing (A+B+C)
			// return

			err = n3c.natsConn.Publish(dispatcherid, msgBytes)
			if err != nil {
				log.Println("write handler unable to publish message: ", err)
				return
			}
		}

	} // mHandler

	// *** set up handler for query by inbound messages ***
	qHandler := func(n3msg *pb.N3Message) (ts []*pb.SPOTuple) {

		// force sender id to be this node
		n3msg.SndId = n3c.pubKey

		tuple := must(messages.DecodeTuple(n3msg.Payload)).(*pb.SPOTuple)
		s, p, o, ctx := tuple.Subject, tuple.Predicate, tuple.Object, n3msg.CtxName
		// fPf("Query : <%s> <%s> <%s> <%s>\n", s, p, o, ctx)

		// check authorisation
		approvalScope := fmt.Sprintf("%s.%s.%s", n3msg.SndId, n3msg.NameSpace, ctx)
		if !n3c.approved(approvalScope) {
			log.Println("you are not authorised to query to context: ", n3msg.NameSpace, ctx)
			return
		}

		// *** <ObjectID / TerminatorID List> query ***
		if IArrEleIn(s, Ss{"*", ""}) && p != "" && o != "" {
			var IDs []string
			switch {
			case p == MARKTerm: //                                     *** <TerminatorID> by Terminator & ObjectID ***
				IDs = dbClient.IDListByPathValue(ctx, p, o, true, false)
			case S(p).Contains(DELIPath): //                           *** <ObjectIDs> by path-value ***
				IDs = dbClient.IDListByPathValue(ctx, p, o, false, true)
			case IArrEleIn(p, Ss{"root", "ROOT", "Root"}): //          *** <ObjectIDs> by root name ***
				IDs = dbClient.IDListByRoot(ctx, o, DELIPath, true)
			}

			if IDs != nil {
				for _, id := range IArrRmRep(Ss(IDs)).([]string) {
					if s == "*" || p == MARKTerm { //                                       *** including deleted ObjectID OR get TerminatorID ***
						ts = append(ts, &pb.SPOTuple{Subject: id, Predicate: p, Object: o})
					} else { //                                                             *** excluding deleted ObjectID ***
						if exist, alive := dbClient.Status(ctx, id); exist && alive {
							ts = append(ts, &pb.SPOTuple{Subject: id, Predicate: p, Object: o})
						}
					}
				}
			}
			return
		}

		// *** <Data> query; <SUBJECT> 's' must be ObjectID / TerminatorID ***
		switch p {
		case MARKTerm: //   *** <Terminator Line's ObjectID> ***
			{
				if _, _, ID, _, found := dbClient.OsBySP(ctx, s, p, false, false, 0, 0); found {
					ts = append(ts, &pb.SPOTuple{Subject: s, Predicate: p, Object: ID[0]})
				}
			}
		case "": //         *** <ROOT> ***
			{
				root := dbClient.RootByID(ctx, s, DELIPath)
				ts = append(ts, &pb.SPOTuple{Subject: s, Predicate: "root", Object: root})
			}
		case "[]", "::": // *** <ARRAY / STRUCT> ***
			{
				metaType := matchAssign(p, "::", "[]", "S", "A").(string)
				if start, end, _ := getVerRange(dbClient, ctx, p+s, metaType); start >= 1 { // *** Meta file to check alive ***
					root := dbClient.RootByID(ctx, s, DELIPath)
					if ss, ps, os, vs, ok := dbClient.OsBySP(ctx, p+s, root, false, true, start, end); ok {
						for i := range ss {
							ts = append(ts, &pb.SPOTuple{Subject: s, Predicate: ps[i], Object: os[i], Version: vs[i]})
						}
					}
				}
			}
		default: //         *** <VALUES> ***
			{
				metaType := trueAssign(S(s).HP("::"), S(s).HP("[]"), "S", "A", "V").(string)

				if S(ctx).HS("-meta") { // *** REQUEST A TICKET FOR PUBLISHING ***
					if ver, ok := mpObjIDVer.Load(s); ok {
						return mkTicket(dbClient, ctx, s, ver.(int64), BIGVER)
					}
					return mkTicket(dbClient, ctx, s, 0, BIGVER)

					// if _, end, v := getVerRange(dbClient, ctx, s, metaType); end != -1 { // *** Meta file to check alive ***
					// 	return mkTicket(dbClient, ctx, s, end, v) //                        *** make a ticket for publishing, -1 means it's dead ***
					// }
					// return
				}

				//          *** GENERAL QUERY ***
				if start, end, _ := getVerRange(dbClient, ctx, s, metaType); start >= 1 { // *** Meta file to check alive ***

					fPln("*** GENERAL QUERY ***")
					// Fetch all tuples [ts]
					dbClient.TuplesBySP(ctx, tuple, &ts, start, end)

					fPln("*** CHECKING PRIVACY ***")
					// DONE: PRIVACY, check privacy rules, modify [ts]
					if !IArrEleIn(ctx, Ss{"ctxid", "privctrl"}) && !S(ctx).HS("meta") {
						for _, pt := range ts {
							if S(pt.Subject).IsUUID() && S(pt.Predicate) != MARKTerm {
								root := sSpl(pt.Predicate, DELIPath)[0]
								if pcCtxPathRW, ok1 := pcObjCtxPathRW[root]; ok1 {
									if pcPathRW, ok2 := pcCtxPathRW[ctx]; ok2 { //              *** has rules ***
										// fPln("DEBUG", pt.Predicate, pcPathRW[pt.Predicate])
										switch {
										case S(pcPathRW[pt.Predicate]).IsUUID():
											continue
										case !IArrEleIn(pcPathRW[pt.Predicate], Ss{"R/W", "R", "RW", "WR"}):
											pt.Object = NOREAD
										}
									}
								}
							}
						}
					}
				}
			}
		}

		return
	} // qHandler

	dHandler := func(n3msg *pb.N3Message) int { //      *** set up handler for delete by inbound messages ***
		return 1234567 //                               *** DO NOT USE THIS TO DELETE, USE 'DEADMARK' IN PUB ***
	}

	// start server
	apiServer := n3grpc.NewAPIServer()
	apiServer.SetMessageHandler(mHandler, qHandler, dHandler)
	return apiServer.Start(viper.GetInt("rpc_port"))
}

//
// checks if sending to the given context is allowed for this user
//
func (n3c *N3Node) approved(scope string) bool {
	n3c.mutex.Lock()
	defer n3c.mutex.Unlock()

	approved, ok := n3c.approvedContexts[scope]
	if !ok {
		return false
	}

	return approved

}

//
// reads from this client's feed and passes to
// storage / query layer
//
func (n3c *N3Node) startReadHandler() error {

	// start up the influx publisher
	// TODO: abstract multi-db handlers to channel reader
	// to multiplex send to different data-stores
	dbClt, err := n3influx.NewDBClient()
	if err != nil {
		return errors.Wrap(err, "read handler cannot connect to influx store:")
	}

	// set up message handler
	lastMessageOffset := viper.GetInt64(n3c.pubKey)
	var nextMessage int64
	if lastMessageOffset != 0 {
		nextMessage = lastMessageOffset + 1
	}
	log.Println("consming messages offset from:", nextMessage)
	handler := func(msg *lbproto.Message, err error) {

		if err != nil {
			log.Println("read handler error from liftbridge server:", err)
			return
		}

		// decode msg from transport format
		n3msg, err := messages.DecodeN3Message(msg.Value)
		if err != nil {
			log.Println("node read handler cannot decode message proto: ", err)
			return
		}
		// log.Println("node received message from:", n3msg.SndId)

		// unpack payload - decrypt and unmarshal
		tuple, err := n3crypto.DecryptTuple(n3msg.Payload, n3msg.DispId, n3c.privKey)
		if err != nil {
			log.Println("read handler decrypt error: ", err)
			return
		}

		// Added ****************************************** //
		s, p, o, _ := tuple.Subject, tuple.Predicate, tuple.Object, tuple.Version

		// PRIVACY *** update pcObjCtxPathRW at runtime ***
		if n3msg.CtxName == "privctrl" {
			forroot = IF(p != MARKTerm, sSpl(p, DELIPath)[0], forroot).(string)
			forctx = IF(S(p).HS("forcontext"), o, forctx).(string)
			if _, ok := pcObjCtxPathRW[forroot]; !ok {
				//                                 context    path   rw
				pcObjCtxPathRW[forroot] = make(map[string]map[string]string)
			}
			if _, ok := pcObjCtxPathRW[forroot][forctx]; !ok { //        *** forctx:  begin with "", then "ctx***" ***
				//                                         path   rw
				pcObjCtxPathRW[forroot][forctx] = make(map[string]string)
			}
			if p == MARKTerm {
				// change forctx("")'s value back to its forctx("abc")'s related value
				for path, rw := range pcObjCtxPathRW[forroot][""] {
					pcObjCtxPathRW[forroot][forctx][path] = IF(forctx != "", rw, "error").(string)
				}
				forctx = ""
				pcObjCtxPathRW[forroot][forctx] = make(map[string]string)

			} else {
				pcObjCtxPathRW[forroot][forctx][p] = o
			}
		}

		// PRIVACY *** delete pcObjCtxPathRW at runtime ***
		if n3msg.CtxName == "ctxid" {
			if o == MARKDelID && pcObjCtxPathRW[s] != nil {
				delete(pcObjCtxPathRW[s], p)
			}
		}

		// Update ObjIDVerCache
		if S(n3msg.CtxName).HS("-meta") {
			mpObjIDVer.Store(s, S(sSpl(o, "-")[1]).ToInt64())
		}

		// only for testing (A+B+D)
		// return

		// *** exclude "legend liftbridge data" ***
		// if inDB(dbClt, n3msg.CtxName, tuple) {
		// 	return
		// }

		err = dbClt.StoreTuple(tuple, n3msg.CtxName)
		if err != nil {
			log.Println("error storing tuple:", err)
			return
		}
		// log.Println("...tuple stored successfully.")
		lastMessageOffset = msg.Offset
	}

	// create context to control async subscription
	ctx, cancelFunc := context.WithCancel(context.Background())
	n3c.addHandlerContext(n3c.pubKey, cancelFunc)

	// create the subscription and run until context is cancelled
	go func() {
		streamName := fmt.Sprintf("%s-stream", n3c.pubKey)
		err := n3c.lbConn.Subscribe(ctx, streamName, handler, liftbridge.StartAtOffset(nextMessage))
		if err != nil {
			log.Println("node error subscribing read handler: ", err)
			n3c.removeHandlerContext(n3c.pubKey)
		}
		<-ctx.Done()
		// store the last read position
		viper.Set(n3c.pubKey, lastMessageOffset)
	}()

	return nil
}

//
// adds the cancel function for context used to control the given handler
//
func (n3c *N3Node) addHandlerContext(name string, cancelFunc func()) {

	n3c.mutex.Lock()
	defer n3c.mutex.Unlock()

	n3c.handlerContexts[name] = cancelFunc

}

//
// invokes the given handler context cancel function, and removes the
// handler from the internal list
//
func (n3c *N3Node) removeHandlerContext(name string) {

	n3c.mutex.Lock()
	defer n3c.mutex.Unlock()

	cfunc, ok := n3c.handlerContexts[name]
	if ok {
		cfunc()
		delete(n3c.handlerContexts, name)
		log.Println("node context handler shut down: ", name)
	}

}

// Close :
// shuts down connections, closes handlers
//
func (n3c *N3Node) Close() {

	// shut down nats connection
	n3c.natsConn.Flush()
	n3c.natsConn.Close()

	// shut down lb connection
	n3c.lbConn.Close()

	// close all handlers by invoking cancelfunc on associated contexts
	for name := range n3c.handlerContexts {
		n3c.removeHandlerContext(name)
	}

	// save the config file, to remember client read position
	err := n3config.SaveConfig()
	if err != nil {
		log.Println("unable to save config:", err)
	}

	log.Println("node successfully shut down")

}
