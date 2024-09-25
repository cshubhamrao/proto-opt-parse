package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	greeterv1 "github.com/cshubhamrao/proto-opt-parse/gen/greeter/v1"
	optsv1 "github.com/cshubhamrao/proto-opt-parse/gen/opts/v1"
	"google.golang.org/protobuf/proto"
)

var log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func main() {
	log.Debug("Starting")
	protoFd := greeterv1.File_greeter_v1_hello_proto // get the file descriptor
	svcs := protoFd.Services()                       // get list of services defined
	for i := 0; i < svcs.Len(); i++ {
		svc := svcs.Get(i) // Get the service
		svcName := svc.FullName()
		fmt.Printf("Got Service --> %s\n", svcName)
		methods := svc.Methods() // List of methods in the service
		for i := 0; i < methods.Len(); i++ {
			method := methods.Get(i) // Get the method
			fmt.Printf("\tMETHOD --> %s\n", method.FullName().Name())

			options := method.Options()                              // Get the options for the method
			if proto.HasExtension(options, optsv1.E_LoggingConfig) { // Check if our extension is present
				extRaw := proto.GetExtension(options, optsv1.E_LoggingConfig) //Get extension as any
				logOpts, ok := extRaw.(*optsv1.LogOptions)                    // cast to our extension's type
				if !ok {
					fmt.Println("Error casting extension")
					continue
				}
				fmt.Printf("\t\tOPTION -> %s\n", logOpts)
				slogLevel := slog.LevelError
				slogLevel.UnmarshalText([]byte(logOpts.GetLogLevel()))
				log.Log(context.Background(), slogLevel, "With Logging Options", "scope", logOpts.GetScopeName())
			}
		}
	}
}
