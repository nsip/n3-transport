// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"
	"strings"

	"../../n3config"
	nats "github.com/nats-io/go-nats"
	"github.com/nsip/n3-messages/messages"
	"github.com/nsip/n3-messages/messages/pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// approveCmd represents the approve command
var approveCmd = &cobra.Command{
	Use:   "approve user(required) contextName(required)",
	Short: "Approve a user for participation in a context",
	Long: `
	Approve a user to participate in a context. Requires arguments:
	user - public-key id of user
	context-name - the name of the context
	for example,
	>./n3cli approve 78CaH3sHwsLHVo5ov6VHPuFRqeiuDsxuhM7spUtB3cEU nsipContext1
	you are only able to approve users for contexts that you own.
	`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("approve called")
		sendApprovals(args)
	},
}

func init() {
	rootCmd.AddCommand(approveCmd)
}

//
// makes the context available to the remote user
// by publishing an approval for the base context
// and the context-meta which holds useful
// strucutral information about the context.
//
func sendApprovals(args []string) {

	// read the config
	err := n3config.ReadConfig()
	if err != nil {
		log.Fatalf("\n\n\tcannot proceed, no valid n3 config found \n\t- run './n3cli init' to create one\n\n")
	}

	// check the user & context
	user := strings.TrimSpace(args[0])
	contextName := strings.TrimSpace(args[1])
	contextNameMeta := contextName + "-meta"
	if user == "" || contextName == "" {
		log.Println("cannot have empty arguments")
		return
	}

	// create approval tuple - predicate "deny" to revoke access
	// approvals are always only for contexts of this user
	approvalTuple := &pb.SPOTuple{Subject: user, Predicate: "grant", Object: contextName}
	tupleBytes, err := messages.EncodeTuple(approvalTuple)
	if err != nil {
		log.Println("unable to encode approval tuple: ", err)
		return
	}
	// log.Printf("%v\n", approvalTuple)

	approvalTupleMeta := &pb.SPOTuple{Subject: user, Predicate: "grant", Object: contextNameMeta}
	tupleBytesMeta, err := messages.EncodeTuple(approvalTupleMeta)
	if err != nil {
		log.Println("unable to encode approval meta tuple: ", err)
		return
	}

	// fetch required identity keys
	mypubKey := viper.GetString("pubkey")

	// context message
	approvalMsg := &pb.N3Message{
		Payload:   tupleBytes,
		SndId:     mypubKey,
		NameSpace: mypubKey,
		CtxName:   "approvals",
	}

	// context-meta message
	approvalMsgMeta := &pb.N3Message{
		Payload:   tupleBytesMeta,
		SndId:     mypubKey,
		NameSpace: mypubKey,
		CtxName:   "approvals",
	}

	approvalBytes, err := messages.EncodeN3Message(approvalMsg)
	if err != nil {
		log.Println("unable to encode approval message: ", err)
	}

	approvalBytesMeta, err := messages.EncodeN3Message(approvalMsgMeta)
	if err != nil {
		log.Println("unable to encode approval -meta message: ", err)
	}

	// send to approve
	natsAddr := viper.GetString("nats_addr")
	nc, err := nats.Connect(natsAddr)
	if err != nil {
		log.Println("cannot connect to nats server: ", err)
		return
	}
	defer nc.Close()

	err = nc.Publish("approvals", approvalBytes)
	if err != nil {
		log.Println("unable to publish approval message: ", err)
		return
	}

	err = nc.Publish("approvals", approvalBytesMeta)
	if err != nil {
		log.Println("unable to publish approval -meta message: ", err)
		return
	}

	log.Printf("\n\n\tContexts created: \n\t %s \n\t %s \n\tContext approvals for user (%s) published to n3 network.\n\n", contextName, contextNameMeta, user)

}
