package selectors

import (
	"agones.dev/agones/pkg/apis"
	allocatorv1 "agones.dev/agones/pkg/apis/allocation/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
)

func GetTowerDefenceSelector(playerCount int64, inProgress bool) *allocatorv1.GameServerAllocation {
	var selectors []allocatorv1.GameServerSelector

	if inProgress {
		selectors = append(selectors, allocatorv1.GameServerSelector{
			LabelSelector: v1.LabelSelector{
				MatchLabels: map[string]string{"agones.dev/fleet": "towerdefence-game", "backfill": "true"},
			},
			GameServerState: &AllocatedState,
			Players: &allocatorv1.PlayerSelector{
				MinAvailable: playerCount,
				MaxAvailable: math.MaxInt64,
			},
		})
	}

	selectors = append(selectors, allocatorv1.GameServerSelector{
		LabelSelector: v1.LabelSelector{
			MatchLabels: map[string]string{"agones.dev/fleet": "towerdefence-game"},
		},
		GameServerState: &ReadyState,
	})

	return &allocatorv1.GameServerAllocation{
		Spec: allocatorv1.GameServerAllocationSpec{
			Scheduling: apis.Packed,
			Selectors:  selectors,
		},
	}
}