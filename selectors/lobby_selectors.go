package selectors

import (
	"agones.dev/agones/pkg/apis"
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	allocatorv1 "agones.dev/agones/pkg/apis/allocation/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
)

var (
	AllocatedState = agonesv1.GameServerStateAllocated
	ReadyState     = agonesv1.GameServerStateReady
)

func GetLobbySelector(playerCount int64) *allocatorv1.GameServerAllocation {
	return &allocatorv1.GameServerAllocation{
		Spec: allocatorv1.GameServerAllocationSpec{
			Scheduling: apis.Packed,
			Selectors: []allocatorv1.GameServerSelector{
				{
					LabelSelector: v1.LabelSelector{
						MatchLabels: map[string]string{"agones.dev/fleet": "lobby"},
					},
					GameServerState: &AllocatedState,
					Players: &allocatorv1.PlayerSelector{
						MinAvailable: playerCount,
						MaxAvailable: math.MaxInt64,
					},
				},
				{
					LabelSelector: v1.LabelSelector{
						MatchLabels: map[string]string{"agones.dev/fleet": "lobby"},
					},
					GameServerState: &ReadyState,
				},
			},
		},
	}
}
