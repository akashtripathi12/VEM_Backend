package handlers

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/akashtripathi12/TBO_Backend/internal/models"
)

func generateTestGuests(count int) []models.Guest {
	guests := make([]models.Guest, count)
	for i := 0; i < count; i++ {
		guests[i] = models.Guest{
			ID:   uuid.New(),
			Name: "TestGuest",
		}
	}
	return guests
}

func TestAIAllocate_TightlyPackedInventory(t *testing.T) {
	// 5 slots, 5 guests (2 families: size 3 and 2)
	eventID := uuid.New()
	fams := []FamilyGroup{
		{FamilyID: "fam1", Size: 3, Guests: generateTestGuests(3), ArrivalDate: time.Now()},
		{FamilyID: "fam2", Size: 2, Guests: generateTestGuests(2), ArrivalDate: time.Now()},
	}

	inv := []VirtualRoom{
		{RoomOfferID: "offer1", RoomName: "Triple Room", MaxCapacity: 3, Available: 1, Total: 1, PricePerRoom: 100},
		{RoomOfferID: "offer2", RoomName: "Double Room", MaxCapacity: 2, Available: 1, Total: 1, PricePerRoom: 80},
	}

	ctx := &OptimisationContext{
		EventID:         eventID,
		UnallocatedFams: fams,
		BaseInventory:   inv,
		OptInventory:    inv,
	}

	runStart := time.Now()
	res := RunHeuristicOptimisation(ctx, runStart)

	if res.TotalWaste != 0 {
		t.Errorf("Expected 0 waste for tightly packed inventory, got %d", res.TotalWaste)
	}
	if len(res.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(res.Suggestions))
	}
}

func TestAIAllocate_ResidualCapacityModelling(t *testing.T) {
	// A room of size 5 has 1 guest allocated. Residual = 4.
	// We want to pack two families of size 2 into it.
	eventID := uuid.New()
	fams := []FamilyGroup{
		{FamilyID: "fam1", Size: 2, Guests: generateTestGuests(2), ArrivalDate: time.Now()},
		{FamilyID: "fam2", Size: 2, Guests: generateTestGuests(2), ArrivalDate: time.Now()},
	}

	inv := []VirtualRoom{
		{RoomOfferID: "offer1", RoomName: "Penthouse (Residual)", MaxCapacity: 4, Available: 1, Total: 1, PricePerRoom: 200, IsVirtualResidual: true},
	}

	ctx := &OptimisationContext{
		EventID:         eventID,
		UnallocatedFams: fams,
		BaseInventory:   inv,
		OptInventory:    inv,
	}

	runStart := time.Now()
	res := RunHeuristicOptimisation(ctx, runStart)

	// Best-Fit Decreasing should place fam1 into the residual room (capacity 4).
	// After that, the room has Available = 0. Standard algorithm doesn't natively "split" a VirtualRoom mid-run.
	// So 1 family gets placed with waste=2, and the other family is unplaceable.
	// Let's assert this behavior.

	if len(res.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(res.Suggestions))
	}
	if res.FamiliesSkipped != 1 {
		t.Errorf("Expected 1 skipped family, got %d", res.FamiliesSkipped)
	}
	if !res.Suggestions[0].IsVirtualResidual {
		t.Errorf("Expected suggestion to be flagged as Virtual Residual")
	}
}

func TestAIAllocate_DeterministicReproducibility(t *testing.T) {
	eventID := uuid.New()
	fams := make([]FamilyGroup, 10)
	for i := 0; i < 10; i++ {
		fams[i] = FamilyGroup{FamilyID: uuid.New().String(), Size: (i % 3) + 1, Guests: generateTestGuests((i % 3) + 1), ArrivalDate: time.Now()}
	}

	inv := []VirtualRoom{
		{RoomOfferID: "offer1", RoomName: "Room A", MaxCapacity: 2, Available: 5, Total: 5, PricePerRoom: 100},
		{RoomOfferID: "offer2", RoomName: "Room B", MaxCapacity: 3, Available: 5, Total: 5, PricePerRoom: 150},
		{RoomOfferID: "offer3", RoomName: "Room C", MaxCapacity: 4, Available: 5, Total: 5, PricePerRoom: 200},
	}

	ctx := &OptimisationContext{
		EventID:         eventID,
		UnallocatedFams: fams,
		BaseInventory:   inv,
		OptInventory:    inv,
	}

	runStart := time.Now()

	// Run twice with exact same time bucket
	res1 := RunHeuristicOptimisation(ctx, runStart)
	res2 := RunHeuristicOptimisation(ctx, runStart)

	if len(res1.Suggestions) != len(res2.Suggestions) {
		t.Errorf("Deterministic runs yielded different lengths: %d != %d", len(res1.Suggestions), len(res2.Suggestions))
	}
	if res1.GlobalScore != res2.GlobalScore {
		t.Errorf("Deterministic runs yielded different scores: %f != %f", res1.GlobalScore, res2.GlobalScore)
	}

	for i := range res1.Suggestions {
		if !reflect.DeepEqual(res1.Suggestions[i], res2.Suggestions[i]) {
			t.Errorf("Mismatch at suggestion %d", i)
		}
	}
}
