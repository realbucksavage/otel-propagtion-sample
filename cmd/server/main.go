package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	otelpropagtionsample "github.com/realbucksavage/otel-propagtion-sample"
	"github.com/realbucksavage/otel-propagtion-sample/generated/greeterpb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func main() {

	{
		instCtx, instClose := context.WithTimeout(context.Background(), 3*time.Second)
		defer instClose()

		closeTraceProvider, err := otelpropagtionsample.InitTracer(instCtx, "greeter-client", "localhost:4317")
		if err != nil {
			panic(err)
		}

		defer closeTraceProvider(instCtx)
	}

	server := grpc.NewServer()
	greeterpb.RegisterGreeterServiceServer(server, &greetingServer{})

	lis, err := net.Listen("tcp", ":4002")
	if err != nil {
		panic(err)
	}

	log.Print("starting grpc server...")
	panic(server.Serve(lis))
}

type greetingServer struct {
	greeterpb.UnimplementedGreeterServiceServer
}

func (server *greetingServer) Greet(ctx context.Context, req *greeterpb.GreetingRequest) (*greeterpb.Greeting, error) {
	var span trace.Span
	ctx, span = otel.Tracer("greeting-server-tracer").Start(ctx, "greet-server")
	defer span.End()

	fakeDelay(ctx)
	return &greeterpb.Greeting{Greeting: fmt.Sprintf("Hello, %s!", req.Name), Language: "English"}, nil
}

func fakeDelay(ctx context.Context) {
	_, span := otel.Tracer("greeting-server-tracer").Start(ctx, "fake-delay")
	defer span.End()

	d := rand.Intn(5)
	span.AddEvent("added delay", trace.WithAttributes(attribute.Int("seconds", d)))

	log.Printf("delaying for %d seconds", d)
	time.Sleep(time.Duration(d) * time.Second)
}
