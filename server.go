package main

import (
	allocatorv1 "agones.dev/agones/pkg/apis/allocation/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/env"
	"log"
	"net"
	"path/filepath"
	"server-discovery-go/proto/service/server_discovery"
	"server-discovery-go/selectors"
)

const (
	namespace = "towerdefence"
	port      = 9090
)

var (
	kubeConfig   = createKubernetesConfig()
	kubeClient   = kubernetes.NewForConfigOrDie(kubeConfig)
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

func createKubernetesConfig() *rest.Config {
	var isInCluster = env.GetString("KUBERNETES_SERVICE_HOST", "") != ""

	var config *rest.Config
	var err error

	if isInCluster {
		config, err = rest.InClusterConfig()
	} else {
		var kubeConfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
	}

	if err != nil {
		panic(err.Error())
	}

	return config
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
	flag.Parse()
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
