package main

import (
	"context"
	"log"

	"./MurmurRPC"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting...")

	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	server := &MurmurRPC.Server{
		Id: proto.Uint32(1),
	}

	go func() {
		streamClient := MurmurRPC.NewV1Client(conn)
		caClient := MurmurRPC.NewV1Client(conn)
		stream, err := streamClient.ServerEvents(context.Background(), server)
		if err != nil {
			log.Fatal(err)
		}
		for {
			event, err := stream.Recv()
			if err != nil {
				break
			}
			if event.GetType() != MurmurRPC.Server_Event_UserConnected {
				continue
			}
			log.Println("Added context action to", event.GetUser().GetName())
			caClient.ContextActionAdd(context.Background(), &MurmurRPC.ContextAction{
				Server:  server,
				Context: proto.Uint32(uint32(MurmurRPC.ContextAction_User)),
				Action:  proto.String("nudge"),
				Text:    proto.String("Nudge!"),
				User:    event.GetUser(),
			})
		}
	}()

	msgClient := MurmurRPC.NewV1Client(conn)
	caClient := MurmurRPC.NewV1Client(conn)
	stream, err := caClient.ContextActionEvents(context.Background(), &MurmurRPC.ContextAction{
		Server: server,
		Action: proto.String("nudge"),
	})
	if err != nil {
		log.Fatal(err)
	}
	for {
		event, err := stream.Recv()
		if err != nil {
			break
		}
		if event.GetUser() == nil {
			continue
		}
		actor, err := msgClient.UserGet(context.Background(), event.GetActor())
		if err != nil {
			continue
		}
		user, err := msgClient.UserGet(context.Background(), event.GetUser())
		if err != nil {
			continue
		}
		if actor.GetSession() == user.GetSession() {
			msgClient.TextMessageSend(context.Background(), &MurmurRPC.TextMessage{
				Server: server,
				Users:  []*MurmurRPC.User{actor},
				Text:   proto.String("Cannot nudge yourself!"),
			})
			continue
		}
		log.Printf("%s nudged %s\n", actor.GetName(), user.GetName())
		msg := &MurmurRPC.TextMessage{
			Server: server,
			Users:  []*MurmurRPC.User{user},
			Text:   proto.String(actor.GetName() + " nudged you!"),
		}
		for i := 1; i <= 3; i++ {
			msgClient.TextMessageSend(context.Background(), msg)
		}
	}
}
