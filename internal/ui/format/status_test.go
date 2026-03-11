package format

import (
	"strings"
	"testing"

	"chronicle-of-a-clan/internal/core/status"
)

func TestStatus(t *testing.T) {
	got := Status(status.View{ClanName: "Chronicle Clan", Day: 7, Gold: 123, Fame: 9, MembersCount: 2, ActiveQuestsCount: 3})
	for _, line := range []string{
		"Clan: Chronicle Clan   Day: 7",
		"Gold: 123   Fame: 9",
		"Members: 2   ActiveQuests: 3",
	} {
		if !strings.Contains(got, line) {
			t.Fatalf("status line missing: %q in %q", line, got)
		}
	}
}
