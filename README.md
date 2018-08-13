# n3-transport
Pluggable transport layer for n3 allowing for stream or blockchain publishing.
Blockchain publisher tagets Tendermint.
Streams will target nats.

A storage publisher for influx is also included.


## Tendermint publisher
Use the github nsip/tendermint fork for the b/c engine.

To publish tuples onto the blockchain:

    import tm "github.com/nsip/n3-transport/n3tendermint"
    import "github.com/nats-io/nuid"
    import "log"
    
    tmpub, err := tm.NewPublisher()
	  if err != nil {
		  panic(err)
	  }

    subject := "4BD6B062-66DD-474B-9E24-E3F85FB61FED" + nuid.Next() //randomise each tuple with nuid
    predicate := "TeachingGroup.TeachingGroupPeriodList.TeachingGroupPeriod[1].DayId"
    object := "F"
    context := "SIF"
    version := 1 // arbitrary for now
    
    msg, err := tmpub.NewMessage(subject, predicate, object, context, version)
    if err != nil {
      panic(err)
    }
    
    err = tmpub.SubmitTx(msg)
    if err != nil {
      log.Println("tx error: ", err)
    } else {
      // log.Println("tx successfully submitted ")
    }
    

## Influx publisher
This component is built into the nsip tendermint instance, but can be used separately, asssumes an instance of influxd server is running and listening on defualt ports:

    import (
      "log"
      
      "github.com/nats-io/nuid"
      inf "github.com/nsip/n3-transport/n3influx"
      "github.com/nsip/n3-transport/pb"
    )

    infpub, err := inf.NewPublisher()
    if err != nil {
      panic(err)
    }
    
    subject := "4BD6B062-66DD-474B-9E24-E3F85FB61FED" + nuid.Next() //randomise each tuple with nuid
    predicate := "TeachingGroup.TeachingGroupPeriodList.TeachingGroupPeriod[1].DayId"
    object := "F"
    context := "SIF"
    version := 1 // arbitrary for now

    err := n3ic.StoreTuple(tuple)
    if err != nil {
      log.println("storage error: ",err)
    }


In the /app area you will find a test-harness example tmpub/pub.go which can be customised to test publish using either service.

The /scripts area contains some useful shell scripts and configs for running tendermint and influx.


    

