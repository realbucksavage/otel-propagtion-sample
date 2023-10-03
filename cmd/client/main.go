package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	otelpropagtionsample "github.com/realbucksavage/otel-propagtion-sample"
	"github.com/realbucksavage/otel-propagtion-sample/generated/greeterpb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	var (
		name   string
		server string
	)
	flag.StringVar(&name, "name", "", "Your name")
	flag.StringVar(&server, "server", ":4002", "Server address")
	flag.Parse()

	{
		instCtx, instClose := context.WithTimeout(context.Background(), 3*time.Second)
		defer instClose()

		closeTraceProvider, err := otelpropagtionsample.InitTracer(instCtx, "greeter-client", "localhost:4317")
		if err != nil {
			panic(err)
		}

		defer closeTraceProvider(instCtx)
	}

	if name == "" {
		panic("--name is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := greeterpb.NewGreeterServiceClient(conn)
	req := &greeterpb.GreetingRequest{Name: name}

	var span trace.Span
	ctx, span = otel.Tracer("greeting-client-tracer").Start(ctx, "greet-client")
	defer span.End()

	resp, err := client.Greet(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		panic(err)
	}

	span.AddEvent("greeting processed")
	fmt.Println(resp)
}
