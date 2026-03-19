package app

import "testing"

func TestSelectedMemberIDsMaintainsListOrder(t *testing.T) {
	m := &model{}
	m.party.memberList = []memberView{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	m.party.selected = map[string]bool{"c": true, "a": true}

	got := m.selectedMemberIDs()
	want := []string{"a", "c"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch: got=%v want=%v", got, want)
		}
	}
}

func TestCurrentAssignmentsReturnsCopy(t *testing.T) {
	m := &model{activeQuestKey: "q", partyByQuest: map[string]partySelection{
		"q": {Assignments: map[string]assignment{"m1": {WeaponID: "W1", ArmorID: "A1"}}},
	}}
	copied := m.currentAssignments()
	copied["m1"] = assignment{WeaponID: "W2", ArmorID: "A2"}

	if m.partyByQuest["q"].Assignments["m1"].WeaponID != "W1" {
		t.Fatalf("source map should not be mutated")
	}
}
