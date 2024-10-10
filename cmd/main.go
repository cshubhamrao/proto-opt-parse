package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	greeterv1 "github.com/cshubhamrao/proto-opt-parse/gen/greeter/v1"
	optsv1 "github.com/cshubhamrao/proto-opt-parse/gen/opts/v1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

var log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func main() {
	log.Debug("Starting")

	protoFds, err := loadProtoDescriptors("./desc.txtpb")
	if err != nil {
		log.Error("Failed to load proto descriptors", "error", err)
		os.Exit(1)
	}

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

			// Extract and print the method comment
			comment := getComment(protoFds, method)
			if comment != "" {
				fmt.Printf("\t\tCOMMENT --> %s\n", comment)
			}

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

func getComment(protoFds *protoregistry.Files, method protoreflect.Descriptor) string {
	md, err := protoFds.FindDescriptorByName(method.FullName())
	if err != nil {
		fmt.Println("Error finding descriptor")
		return ""
	}
	return strings.TrimSpace(md.ParentFile().SourceLocations().ByDescriptor(md).LeadingComments)
}

func loadProtoDescriptors(filePath string) (*protoregistry.Files, error) {
	descriptorBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor file: %w", err)
	}

	fileDescriptorSet := &descriptorpb.FileDescriptorSet{}
	if err := prototext.Unmarshal(descriptorBytes, fileDescriptorSet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptor: %w", err)
	}

	return protodesc.NewFiles(fileDescriptorSet)
}
