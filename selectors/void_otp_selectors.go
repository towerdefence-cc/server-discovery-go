package selectors

import (
	"agones.dev/agones/pkg/apis"
	allocatorv1 "agones.dev/agones/pkg/apis/allocation/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
)

func GetVoidOtpSelector() *allocatorv1.GameServerAllocation {
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
						MinAvailable: 1,
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
