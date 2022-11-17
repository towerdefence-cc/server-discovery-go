package main

import (
	"agones.dev/agones/pkg/apis"
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
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
	"math"
	"net"
	"path/filepath"
	"server-discovery-go/proto/service/server_discovery"
)

const (
	namespace = "towerdefence"
	port      = 9090
)

var (
	kubeConfig   = createKubernetesConfig()
	kubeClient   = kubernetes.NewForConfigOrDie(kubeConfig)
	agonesClient = versioned.NewForConfigOrDie(kubeConfig)

	allocatedState = agonesv1.GameServerStateAllocated
	readyState     = agonesv1.GameServerStateReady
)

type serverDiscoveryServer struct {
	server_discovery.UnimplementedServerDiscoveryServer
}

func (s *serverDiscoveryServer) GetSuggestedLobbyServer(ctx context.Context, _ *empty.Empty) (*server_discovery.LobbyServer, error) {
	allocationRequest := &allocatorv1.GameServerAllocation{
		Spec: allocatorv1.GameServerAllocationSpec{
			Scheduling: apis.Packed,
			Selectors: []allocatorv1.GameServerSelector{
				{
					LabelSelector: v1.LabelSelector{
						MatchLabels: map[string]string{"agones.dev/fleet": "lobby"},
					},
					GameServerState: &allocatedState,
					Players: &allocatorv1.PlayerSelector{
						MinAvailable: 1,
						MaxAvailable: math.MaxInt64,
					},
				},
				{
					LabelSelector: v1.LabelSelector{
						MatchLabels: map[string]string{"agones.dev/fleet": "lobby"},
					},
					GameServerState: &readyState,
				},
			},
		},
	}

	allocation, err := agonesClient.AllocationV1().GameServerAllocations(namespace).Create(ctx, allocationRequest, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	allocationState := allocation.Status.State

	if allocationState == "UnAllocated" {
		return nil, status.Error(codes.NotFound, "No available lobby servers")
	}

	allocationStatus := allocation.Status
	return &server_discovery.LobbyServer{
		ConnectableServer: &server_discovery.ConnectableServer{
			Id:      allocationStatus.GameServerName,
			Address: allocationStatus.Address,
			Port:    uint32(allocationStatus.Ports[0].Port),
		},
	}, nil
}

func (s *serverDiscoveryServer) GetSuggestedOtpServer(ctx context.Context, _ *empty.Empty) (*server_discovery.ConnectableServer, error) {
	allocationRequest := &allocatorv1.GameServerAllocation{
		Spec: allocatorv1.GameServerAllocationSpec{
			Scheduling: apis.Packed,
			Selectors: []allocatorv1.GameServerSelector{
				{
					LabelSelector: v1.LabelSelector{
						MatchLabels: map[string]string{"agones.dev/fleet": "void-otp"},
					},
					GameServerState: &allocatedState,
					Players: &allocatorv1.PlayerSelector{
						MinAvailable: 1,
						MaxAvailable: math.MaxInt64,
					},
				},
				{
					LabelSelector: v1.LabelSelector{
						MatchLabels: map[string]string{"agones.dev/fleet": "void-otp"},
					},
					GameServerState: &readyState,
				},
			},
		},
	}

	allocation, err := agonesClient.AllocationV1().GameServerAllocations(namespace).Create(ctx, allocationRequest, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	allocationState := allocation.Status.State

	if allocationState == "UnAllocated" {
		return nil, status.Error(codes.NotFound, "No available lobby servers")
	}

	allocationStatus := allocation.Status
	return &server_discovery.ConnectableServer{
		Id:      allocationStatus.GameServerName,
		Address: allocationStatus.Address,
		Port:    uint32(allocationStatus.Ports[0].Port),
	}, nil
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
