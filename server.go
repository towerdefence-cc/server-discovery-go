package main

import (
	allocatorv1 "agones.dev/agones/pkg/apis/allocation/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/towerdefence-cc/go-utils/kubeutils"
	"github.com/towerdefence-cc/grpc-api-specs/gen/go/service/server_discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"log"
	"net"
	"server-discovery-go/selectors"
)

const (
	namespace = "towerdefence"
	port      = 9090
)

var (
	kubeConfig   = createKubernetesClient()
	agonesClient = versioned.NewForConfigOrDie(kubeConfig)
)

type serverDiscoveryServer struct {
	server_discovery.UnimplementedServerDiscoveryServer
}

func (s *serverDiscoveryServer) GetSuggestedLobbyServer(ctx context.Context, request *server_discovery.ServerRequest) (*server_discovery.LobbyServer, error) {
	allocation, err := agonesClient.AllocationV1().GameServerAllocations(namespace).Create(ctx, selectors.GetLobbySelector(request.PlayerCount), v1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	connectableServer, err := createConnectableServer("lobby", allocation.Status)
	if err != nil {
		return nil, err
	}

	return &server_discovery.LobbyServer{
		ConnectableServer: connectableServer,
	}, nil
}

func (s *serverDiscoveryServer) GetSuggestedOtpServer(ctx context.Context, _ *empty.Empty) (*server_discovery.ConnectableServer, error) {
	allocation, err := agonesClient.AllocationV1().GameServerAllocations(namespace).Create(ctx, selectors.GetVoidOtpSelector(), v1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createConnectableServer("void-otp", allocation.Status)
}

func (s *serverDiscoveryServer) GetSuggestedTowerDefenceServer(ctx context.Context, request *server_discovery.TowerDefenceServerRequest) (*server_discovery.ConnectableServer, error) {
	allocation, err := agonesClient.AllocationV1().GameServerAllocations(namespace).
		Create(ctx, selectors.GetTowerDefenceSelector(1, request.GetInProgress()), v1.CreateOptions{})

	if err != nil {
		log.Print(err)
		return nil, err
	}

	return createConnectableServer("tower-defence-game", allocation.Status)
}

func createConnectableServer(fleetType string, allocation allocatorv1.GameServerAllocationStatus) (*server_discovery.ConnectableServer, error) {
	if allocation.State == "UnAllocated" {
		return nil, status.Errorf(codes.NotFound, "No available %s servers", fleetType)
	}

	return &server_discovery.ConnectableServer{
		Id:      allocation.GameServerName,
		Address: allocation.Address,
		Port:    uint32(allocation.Ports[0].Port),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	server_discovery.RegisterServerDiscoveryServer(server, &serverDiscoveryServer{})

	log.Printf("Starting server at %s", lis.Addr().String())
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}

func createKubernetesClient() *rest.Config {
	config, err := kubeutils.CreateKubernetesConfig()
	if err != nil {
		panic(err)
	}
	return config
}
