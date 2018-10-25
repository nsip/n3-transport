### This repo under heavy development, and this readme is for internal use primarily. If you're interested in n3 please get in touch via nsip.edu.au.


## Getting Started
N3 Transport is a collection of components and services to support education data sharing:

##### n3node
The core binary, presents a grpc interface to accept tuples for publishing to the sharing network, and manages the publishing of data and the receipt of data from the network.
Nodes have a public-key based identity, and encrypt and sign all data submitted through them onto the network.
Nodes connect to either a streaming engine or a blockchain engine in order to share data with other nodes on the network.
When nodes receive data from the read stream they pass it on for storage to any of the supported storage engines, the distribution ships with influx as a local data store.

##### n3cli
A command-line interface tool (n3cli) is provided as part of the distribution to allow management and configuraiton functions to be applied to the node and the network.

init - creates the public-key identity of the node

context create - all data must be associated with a context, even if the data is of different types. Contexts are owned by their creator.

approve user context - grants access by another user/node to the data within a context. When you create a context you are autmatically granted approval as the owner.
Data published without a context or to a context for which the user is not approved is simply discarded by the transport layer.

(WIP) Futher commands will be added to support dynamic addition of context and user-specific privacy restrictions.

#### services
##### dispatcher
Nodes never communicate to one another directly, nodes are assinged to a dispatcher and data is published to encrypted streams, dispatchers propagate data from those anonymous streams to the read streams of approved nodes, and data is encrypted so that it can only be read by the recipient.

##### nats/liftbridge
The streaming/messaing engine that allows all components to communicate, can be deployed locally, centrally, clustered or peer to peer.

##### tendermint
Alternative node transport where communication is between blockchain nodes; implicitly peer to peer. Access and consensus are managed by the configuration ownership and voting rights set for the blockchain network.

##### influx
Influx is the default data store provided as part of the distribution, the core service is the influxd daemon which is launched on node startup, influx the database's own cli for data exploration is also included

#### test components
/n3-transport/app/test contains a number of testing utilities for exercising different parts of n3.
The most useful is probably n3pub which can be used to send test batches of data through the infrastrucure.

### Build & Run
to fetch everything, normal go get:

`go get -u github.com/nsip/n3-transport`

to create a working distribution you'll need to run the build script (mac only currently, easily changed for linux but we currently have one external dependency not building for windows)

```>./build.sh```

this will create a /build folder, cd to that and then run the following commands using the ./n3cli tool

```>./n3cli init```

this establishes the identity for this node and creates a config file with default connection params for all services.

```>./n3cli create context [context-name]```

create a context to publish data to, without a context data will not be sent to the network.

Launch the n3 node:

```>./n3node```

this will start the node and all associated services.
There are command-line flags for n3node (available with `./n3node --help`):

--withServices (defaults to true); launches all services as well as the node. Set to false if you want to just start another node acting as a consumer against shared infrastrucure.

--servicesOnly (defaults to false); launches just the supporting services bundle.

The streaming, dispatcher and influx services will be started.

The node will announce it's id, connect to the network and be assigned a dispatcher and start up the grpc API handler to accept incoming data. It will also start a storage handler to write any received data to influx.








## n3-transport
N3 transport is a collection of services and components to allow the collection, distribution and analysis of educaiton data-sets.

N3 models data exchange as a synchronisation between nodes. The synchronisation services can be deployed locally, as a centalised cluster, or in peer-to-peer fashion between collaborating nodes.

Nodes themselves can act as a gateway for an individual with a node running on a local machine or local network, or be used in a centralised fashion with multiple clients sending data to a central node.

Nodes maintain separation between read & write activities. Nodes communicate through either a stream-based protocol or a blockchain protocol (tendermint) depending on the levels of security, auditability and consesnus required by the operaitonal domain.

Data is published to the network of participants by sending data into a node, either as files, as rest-html calls, or using the node's gRPC API.

Published data is written to encrypted streams or the blockchain for consumption by other participants. In the streaming model access to data is controlled by approvals granted explicitly by data owners, on the blockchain access is goverend by the consensus and voting regime established for the blockchain.

All data traversing the network is encrypted and signed using nacl-box. To enforce this individual nodes never publish to one another but to independnent encrypted streams from where infrastrucure actors (dispatchers) enforce privacy constraints and re-encrypt data for approved recipients feeding it to the read stream of those approved nodes.

Data is read from the network by nodes reading from their private encrypted stream, and data is then stored into a variety of data stores, with InfluxDB being the default store provided with the distribution. As with all n3 services influx can be run locally to the node or deployed centrally, the distribution contains a local instance.

The use of persistent streams for both publishing and reading data means that nodes can leave the network at any time, and upon rejoining will receive all unseen messages until the local node is up to date with all other nodes in the network.

## data formats
N3 can receive data in xml, json and tuple formats. Data in structured formats such as xml or json is decomposed into its atomic tuples (subject-predicate-object -> entity-attribute-value), and an internal lamport version is assigned across the network as the data is published.
Data once in the data store can be re-inflated into its orginal format, or can be exploited as a graph to link data via any of the tuple relatinships.
The query layer for data exploration is not part of the n3-transport package, but a separate GraphQL-based service available as part of the n3 family (WIP).
Within the eduction domain, then, users can freely mix SIF, EdFI, xAPI, Caliper, LTI etc. data with it all being managed and secured across the transport layer.

## WIP, please contact us if you want to know more about the project, this readme will be updated more thoroughly as work continues
we may even go crazy and include a diagram at some point!

