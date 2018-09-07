package n3tendermint

import (
	"log"

	"github.com/tendermint/tendermint/rpc/client"
)

type Publisher struct {
	bcp *BlockChainPublisher
}

func NewPublisher() (*Publisher, error) {
	n3c := &Publisher{}
	bcClient, err := NewBlockChainPublisher("")
	if err != nil {
		return nil, err
	}
	n3c.bcp = bcClient
	return n3c, nil
}

func (n3c *Publisher) SubmitTx(tx []byte) error {
	bres, err := n3c.bcp.client.BroadcastTxSync(tx)
	if err != nil {
		log.Println("tx error: bc server response: ", bres)
		log.Printf("\n%v\n", bres)
		return err
	}
	return nil
}

type BlockChainPublisher struct {
	client *client.HTTP
}

func NewBlockChainPublisher(port string) (*BlockChainPublisher, error) {
	if port == "" {
		port = "26657"
	}
	bcp := &BlockChainPublisher{
		client: createBlockChainClient(port),
	}

	// test ping to see if connection is ok
	log.Println("test ping to bc server...")
	_, err := bcp.client.Health()
	if err != nil {
		log.Println("error reaching bc server, check it's running and rpc port is correct")
		return nil, err
	}
	log.Println("...result: ping OK")
	return bcp, nil
}

func createBlockChainClient(port string) *client.HTTP {
	rpcAddr := "tcp://0.0.0.0:" + port
	return client.NewHTTP(rpcAddr, "/websocket")
}
