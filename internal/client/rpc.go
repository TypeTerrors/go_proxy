package rpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/lipgloss"
	"google.golang.org/grpc"

	pb "prx/proto" // Adjust to your project's import path.
)

// Define common Lip Gloss styles.
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575")). // green
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555")). // red
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("240")). // gray
			Padding(0, 1)
)

// Client represents your RPC client.
type Client struct {
	Host string
}

func (c *Client) Call(req any) {
	switch v := req.(type) {
	case *pb.AddNewProxyRequest:
		c.AddNewProxy(v)
	case *pb.DeleteProxyRequest:
		c.DeleteProxy(v)
	case *pb.PatchProxyRequest:
		c.PatchProxy(v)
	case *pb.GetRedirectionRecordsRequest:
		c.GetRedirectionRecords(v)
	case *pb.HealthRequest:
		c.Health()
	default:

	}
}

// formatAndPrintResponse formats and prints the response using Lip Gloss styles.
// It accepts a method name (as a string) and a response of any type.
func (c *Client) formatAndPrintResponse(methodName string, response interface{}) {
	// Build a common header for every response.
	header := headerStyle.Render(fmt.Sprintf("[ %s ]", methodName))

	switch res := response.(type) {
	case *pb.Response:
		// For common responses that use pb.Response (e.g., AddNewProxy, DeleteProxy, PatchProxy)
		if res.Success {
			fmt.Println(header + " " + successStyle.Render("Success") + ": Operation completed successfully.")
		} else {
			fmt.Println(header + " " + errorStyle.Render("Error") + ": " + res.Error)
		}
	case *pb.GetRedirectionRecordsResponse:
		// Format a list of redirection records.
		var recordsStr string
		if len(res.GetRecords()) == 0 {
			recordsStr = "No redirection records found."
		} else {
			for _, rec := range res.GetRecords() {
				recordsStr += fmt.Sprintf("From: %s → To: %s\n", rec.GetFrom(), rec.GetTo())
			}
		}
		fmt.Println(header)
		fmt.Println(infoStyle.Render(recordsStr))
	case *pb.HealthResponse:
		// Format a health check response.
		healthInfo := fmt.Sprintf("Status: %s, Time: %s, Version: %s", res.GetStatus(), res.GetTime(), res.GetVersion())
		fmt.Println(header)
		fmt.Println(infoStyle.Render(healthInfo))
	default:
		// Fallback for any unhandled types.
		fmt.Println(header + " " + infoStyle.Render("Unrecognized response type."))
	}
}

// Sample method to call AddNewProxy on the gRPC server and format its response.
func (c *Client) AddNewProxy(req *pb.AddNewProxyRequest) {
	// Dial the gRPC server; adjust the address and port as needed.
	conn, err := grpc.NewClient(c.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewProxyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.AddNewProxy(ctx, req)
	if err != nil {
		// Format the error output with Lip Gloss.
		fmt.Println(errorStyle.Render("RPC call failed: " + err.Error()))
		return
	}
	c.formatAndPrintResponse("AddNewProxy", res)
}

// Similarly, other methods on *Client can use formatAndPrintResponse to handle output.
func (c *Client) DeleteProxy(req *pb.DeleteProxyRequest) {
	conn, err := grpc.NewClient(c.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewProxyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.DeleteProxy(ctx, req)
	if err != nil {
		fmt.Println(errorStyle.Render("RPC call failed: " + err.Error()))
		return
	}
	c.formatAndPrintResponse("DeleteProxy", res)
}

func (c *Client) PatchProxy(req *pb.PatchProxyRequest) {
	conn, err := grpc.NewClient(c.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewProxyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.PatchProxy(ctx, req)
	if err != nil {
		fmt.Println(errorStyle.Render("RPC call failed: " + err.Error()))
		return
	}
	c.formatAndPrintResponse("PatchProxy", res)
}

func (c *Client) GetRedirectionRecords(req *pb.GetRedirectionRecordsRequest) {
	conn, err := grpc.NewClient(c.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewProxyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.GetRedirectionRecords(ctx, req)
	if err != nil {
		fmt.Println(errorStyle.Render("RPC call failed: " + err.Error()))
		return
	}
	c.formatAndPrintResponse("GetRedirectionRecords", res)
}

func (c *Client) Health() {
	conn, err := grpc.NewClient(c.Host, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewProxyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Health(ctx, &pb.HealthRequest{})
	if err != nil {
		fmt.Println(errorStyle.Render("RPC call failed: " + err.Error()))
		return
	}
	c.formatAndPrintResponse("Health", res)
}
