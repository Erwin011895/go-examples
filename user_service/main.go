package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/Erwin011895/go-examples/user_service/proto/pb"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

type User struct {
	Id   string
	Name string
}

var Users = []User{
	{"339c7bb1-6951-430c-8813-8f967536727d", "Erwin"},
	{"4e99b2d8-74b0-4be0-bb65-4b0822634bbe", "Irfan"},
	{"ae3806cb-0f20-4104-ab5d-fc28c7da6eb1", "Kornel"},
	{"c9293056-1d15-4cea-8e79-8773aae6cb00", "Jody"},
}

type server struct {
	pb.UnimplementedUserServiceServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *server) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.User, error) {
	log.Printf("Received: %v", in.GetId())
	for _, u := range Users {
		if u.Id == in.Id {
			return &pb.User{
				Id:   u.Id,
				Name: u.Name,
			}, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *server) NotifAllUser(ctx context.Context, in *pb.NotifAllUserRequest) (*pb.HelloReply, error) {
	log.Println("Start grpc NotifAllUser", in.NotifMessage)
	produceNotifAllUser(ctx, in)
	log.Println("End grpc NotifAllUser")
	return &pb.HelloReply{Message: "success"}, nil
}

func produceNotifAllUser(ctx context.Context, in *pb.NotifAllUserRequest) error {
	// to produce messages
	topic := "NOTIF_ALL_USERS"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte(in.NotifMessage)},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
	log.Println("NOTIF_ALL_USERS produced", in.NotifMessage)
	return err
}

func main() {
	modePtr := flag.String("m", "grpc", "grpc | consumer")
	flag.Parse()

	log.Println(*modePtr)
	switch *modePtr {
	case "grpc":
		runAsGrpcServer()
	case "consumer":
		runAsKafkaConsumer()
	default:
		log.Println("no mode selected")
	}
}

func runAsGrpcServer() {
	lis, err := net.Listen("tcp", ":8880")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runAsKafkaConsumer() {
	log.Println("consumer start")
	topic := "NOTIF_ALL_USERS"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	// handle stop consumer with signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	doneCh := make(chan bool)

	// run inf-loop in goroutine
	go func() {
		dataCh := make(chan string, 1)
		defer close(dataCh)
		b := make([]byte, 10e3)
		for {
			n, err := conn.Read(b)
			if err != nil {
				log.Println(err)
			}
			if n > 0 {
				dataCh <- string(b[:n])
			}

			select {
			case msg := <-dataCh:
				// logic to notify all users
				for _, u := range Users {
					log.Println("notify", u.Name, ": ", msg)
					time.Sleep(5 * time.Second) // simulate long running task
				}
				b = make([]byte, 10e3)
			case <-signals:
				log.Println("Interrupt is detected")
				doneCh <- true
			}
		}
	}()

	<-doneCh // block until signal send

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err)
	}
	log.Println("consumer stopped")
}
