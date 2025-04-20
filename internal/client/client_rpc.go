package client

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"prx/internal/pb"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Run(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <add|update|delete|list> [flags]\n", os.Args[0])
		os.Exit(1)
	}
	subcmd := args[0]

	var err error
	var addr, token, from, to, certPath, keyPath *string
	switch subcmd {
	case "add", "update":
		fs := flag.NewFlagSet(subcmd, flag.ExitOnError)
		addr = fs.String("addr", os.Getenv("PROXY_HOST"), "gRPC server address")
		token = fs.String("token", "", "JWT bearer token")
		from = fs.String("from", "", "source host")
		to = fs.String("to", "", "target URL")
		certPath = fs.String("cert", "", "path to TLS cert")
		keyPath = fs.String("key", "", "path to TLS key")
		fs.Parse(args[1:])
	case "delete":
		fs := flag.NewFlagSet(subcmd, flag.ExitOnError)
		addr = fs.String("addr", os.Getenv("PROXY_HOST"), "gRPC server address")
		token = fs.String("token", "", "JWT bearer token")
		from = fs.String("from", "", "source host")
		fs.Parse(args[1:])
	case "list":
		fs := flag.NewFlagSet(subcmd, flag.ExitOnError)
		addr = fs.String("addr", os.Getenv("PROXY_HOST"), "gRPC server address")
		token = fs.String("token", "", "JWT bearer token")
		fs.Parse(args[1:])
	case "help":
		PrintHelp()
		os.Exit(1)
	default:
		PrintHelp()
		os.Exit(1)
	}

	var missing []string
	if (subcmd == "add" || subcmd == "update") && (*from == "" || *to == "" || *certPath == "" || *keyPath == "" || *token == "") {
		if *from == "" {
			missing = append(missing, "from")
		}
		if *to == "" {
			missing = append(missing, "to")
		}
		if *certPath == "" {
			missing = append(missing, "cert")
		}
		if *keyPath == "" {
			missing = append(missing, "key")
		}
		if *token == "" {
			missing = append(missing, "token")
		}
	} else if subcmd == "delete" && (*from == "" || *token == "") {
		if *from == "" {
			missing = append(missing, "from")
		}
		if *token == "" {
			missing = append(missing, "token")
		}
	} else if subcmd == "list" && *token == "" {
		missing = append(missing, "token")
	}

	if len(missing) > 0 {
		fmt.Printf("Error: missing required flags: %s\n", strings.Join(missing, ", "))
		PrintHelp()
		os.Exit(1)
	}

	conn, err := grpc.NewClient(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewReverseClient(conn)

	baseCtx := metadata.AppendToOutgoingContext(context.Background(),
		"Authorization", "Bearer "+*token)
	ctx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel()

	switch strings.ToLower(subcmd) {
	case "add", "update":
		certBytes, _ := os.ReadFile(*certPath)
		keyBytes, _ := os.ReadFile(*keyPath)
		req := &pb.ProxyRequest{
			From: *from,
			To:   *to,
			Cert: base64.StdEncoding.EncodeToString(certBytes),
			Key:  base64.StdEncoding.EncodeToString(keyBytes),
		}
		if subcmd == "add" {
			_, err = client.Add(ctx, req)
		} else {
			_, err = client.Update(ctx, req)
		}

	case "delete":
		_, err = client.Delete(ctx, &pb.DeleteRequest{From: *from})

	case "list":
		resp, err := client.List(ctx, &pb.ListRequest{})
		if err == nil {
			for _, r := range resp.Records {
				log.Printf("%s -> %s\n", r.From, r.To)
			}
		}
	}

	if err != nil {
		log.Fatal(err)
	}
	log.Println("OK")
}
