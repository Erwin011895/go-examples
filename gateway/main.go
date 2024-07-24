package main

import (
	"context"
	"log"
	"time"

	"github.com/Erwin011895/go-examples/gateway/proto/pb"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var userServiceClient pb.UserServiceClient

func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient("localhost:8880", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	userServiceClient = pb.NewUserServiceClient(conn)

	// Set up http server
	app := fiber.New()
	app.Get("/say-hello", SayHello)
	app.Get("/notif", NotifAllUser)
	log.Fatal(app.Listen(":3000"))
}

func SayHello(c *fiber.Ctx) error {
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := userServiceClient.SayHello(ctx, &pb.HelloRequest{Name: "Erwin"})
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	return c.SendString(r.GetMessage())
}

func NotifAllUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := userServiceClient.NotifAllUser(ctx, &pb.NotifAllUserRequest{NotifMessage: "QWERTY"})
	if err != nil {
		log.Fatalf("%+v", err)
	}

	return c.SendString(r.GetMessage())
}
