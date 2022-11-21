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
		// In game and backfilling
		selectors = append(selectors, allocatorv1.GameServerSelector{
			LabelSelector: v1.LabelSelector{
				MatchLabels: map[string]string{"agones.dev/fleet": "tower-defence-game", "agones.dev/sdk-phase": "game", "agones.dev/sdk-backfill": "true"},
			},
			GameServerState: &AllocatedState,
			Players: &allocatorv1.PlayerSelector{
				MinAvailable: playerCount,
				MaxAvailable: math.MaxInt64,
			},
		})
	}

	// in lobby and backfilling
	selectors = append(selectors, allocatorv1.GameServerSelector{
		LabelSelector: v1.LabelSelector{
			MatchLabels: map[string]string{"agones.dev/fleet": "tower-defence-game", "agones.dev/sdk-phase": "lobby", "agones.dev/sdk-backfill": "true"},
		},
		GameServerState: &AllocatedState,
		Players: &allocatorv1.PlayerSelector{
			MinAvailable: playerCount,
			MaxAvailable: math.MaxInt64,
		},
	})

	// new game server
	selectors = append(selectors, allocatorv1.GameServerSelector{
		LabelSelector: v1.LabelSelector{
			MatchLabels: map[string]string{"agones.dev/fleet": "tower-defence-game"},
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
