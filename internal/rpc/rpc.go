package rpc

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"prx/internal/pb"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		token = fs.String("token", os.Getenv("PROXY_TOKEN"), "JWT bearer token")
		from = fs.String("from", "", "source host")
		to = fs.String("to", "", "target URL")
		certPath = fs.String("cert", "", "path to TLS cert")
		keyPath = fs.String("key", "", "path to TLS key")
		fs.Parse(args[1:])
	case "delete":
		fs := flag.NewFlagSet(subcmd, flag.ExitOnError)
		addr = fs.String("addr", os.Getenv("PROXY_HOST"), "gRPC server address")
		token = fs.String("token", os.Getenv("PROXY_TOKEN"), "JWT bearer token")
		from = fs.String("from", "", "source host")
		fs.Parse(args[1:])
	case "list":
		fs := flag.NewFlagSet(subcmd, flag.ExitOnError)
		addr = fs.String("addr", os.Getenv("PROXY_HOST"), "gRPC server address")
		token = fs.String("token", os.Getenv("PROXY_TOKEN"), "JWT bearer token")
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

	host := strings.Split(*addr, ":")[0]
	tlsCfg := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	creds := credentials.NewTLS(tlsCfg)
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal("failed to dial gRPC server:", "err", err)
	}
	defer conn.Close()

	client := pb.NewReverseClient(conn)
	baseCtx := metadata.AppendToOutgoingContext(context.Background(),
		"Authorization", "Bearer "+*token)
	ctx, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel()

	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("34"))
	infoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63"))

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
		var action string
		if subcmd == "add" {
			_, err = client.Add(ctx, req)
			action = "Added"
		} else {
			_, err = client.Update(ctx, req)
			action = "Updated"
		}
		if err != nil {
			log.Fatal(fmt.Sprintf("%s redirection record failed:", action), "err", err)
		}
		fmt.Println("")
		header := fmt.Sprintf("%s redirection record:", action)
		fmt.Println(infoStyle.Render(header))
		fmt.Printf("%s  %s\n",
			lipgloss.NewStyle().Bold(true).Render("FROM:"), *from)
		fmt.Printf("%s  %s\n\n",
			lipgloss.NewStyle().Bold(true).Render("TO:"), *to)
		fmt.Println("")

	case "delete":
		_, err = client.Delete(ctx, &pb.DeleteRequest{From: *from})
		if err != nil {
			log.Fatal("Delete redirection record failed:", "err", err)
		}
		fmt.Println("")
		header := "Deleted redirection record:"
		fmt.Println(successStyle.Render(header))
		fmt.Printf("%s  %s\n\n",
			lipgloss.NewStyle().Bold(true).Render("FROM:"), *from)
		fmt.Println("")

	case "list":
		resp, err := client.List(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatal("Error retrieving list of redirection records:", "err", err)
		}
		if len(resp.Records) < 1 {
			log.Info("No records found")
		} else {
			maxFrom, maxTo := len("FROM"), len("TO")
			for _, r := range resp.Records {
				if len(r.From) > maxFrom {
					maxFrom = len(r.From)
				}
				if len(r.To) > maxTo {
					maxTo = len(r.To)
				}
			}

			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("63"))
			cellStyle := lipgloss.NewStyle().
				PaddingRight(1)
			fmt.Println("")
			fmt.Printf("%s  %s\n",
				headerStyle.Render(fmt.Sprintf("%-*s", maxFrom, "FROM")),
				headerStyle.Render(fmt.Sprintf("%-*s", maxTo, "TO")),
			)

			sepFrom := strings.Repeat("─", maxFrom)
			sepTo := strings.Repeat("─", maxTo)
			fmt.Printf("%s  %s\n", sepFrom, sepTo)

			for _, r := range resp.Records {
				fmt.Printf("%s  %s\n",
					cellStyle.Render(fmt.Sprintf("%-*s", maxFrom, r.From)),
					cellStyle.Render(fmt.Sprintf("%-*s", maxTo, r.To)),
				)
			}
			fmt.Println("")
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}
