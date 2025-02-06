package main

import (
	context "context"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	callBackFunc func()
	UnimplementedDiscoveryHandlerServer
}

const ServerSocket = "/var/lib/akri/rtsp.sock"
const AkriRegistrationSocket = "unix:/var/lib/akri/agent-registration.sock"
const DiscoveryName = "rtsp"

func main() {
	os.Remove(ServerSocket)

	s := grpc.NewServer()
	listener, err := net.Listen("unix", ServerSocket)
	if err != nil {
		panic(err)
	}
	var server = &Server{}
	server.callBackFunc = func() {
		register(context.Background(), AkriRegistrationSocket)
	}

	RegisterDiscoveryHandlerServer(s, server)
	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	register(context.Background(), AkriRegistrationSocket)
	log.Println("listening")
	select {}
}

const CheckInterval = time.Second * 10 // 10s

func (s *Server) Discover(req *DiscoverRequest, resp DiscoveryHandler_DiscoverServer) error {
	endpoint := req.GetDiscoveryDetails()
	ticker := time.NewTicker(CheckInterval)
	defer func() {
		if s.callBackFunc != nil {
			s.callBackFunc()
		}
	}()

	for {
		select {
		case <-resp.Context().Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			devices := discoverDevices(endpoint)
			if err := resp.Send(&DiscoverResponse{
				Devices: devices,
			}); err != nil {
				return err
			}
		}
	}
}

func discoverDevices(endpoint string) []*Device {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil
	}

	if !strings.Contains(endpoint, "://") {
		endpoint = "schema://" + endpoint
	}

	url, err := url.Parse(endpoint)
	if err != nil {
		return nil
	}

	conn.Close()
	// Add more devices as needed
	return []*Device{
		{
			Id: endpoint,
			Properties: map[string]string{
				"Protocol": "RTSP",
				"Endpoint": url.Hostname(),
				"Port":     url.Port(),
			},
		},
	}
}

func register(ctx context.Context, endpoint string) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		panic(err)
	}

	var client = NewRegistrationClient(conn)
	_, err = client.RegisterDiscoveryHandler(ctx, &RegisterDiscoveryHandlerRequest{
		Name:         DiscoveryName,
		Endpoint:     ServerSocket,
		EndpointType: RegisterDiscoveryHandlerRequest_UDS,
		Shared:       true,
	})
	if err != nil {
		panic(err)
	}
}
