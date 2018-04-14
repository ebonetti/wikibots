package wikibots

import (
	"context"
	"testing"
)

func TestUnit(t *testing.T) {
	name, enID, itID := "ClueBot NG", uint32(13286072), uint32(1598385)
	enID2Name, enErr := New(context.Background(), "en")
	itID2Name, itErr := New(context.Background(), "it")
	switch {
	case enErr != nil:
		t.Error("New returns the following error", enErr)
	case itErr != nil:
		t.Error("New returns the following error", itErr)
	}
	enName := enID2Name[enID]
	itName := itID2Name[itID]
	switch {
	case enName != name:
		t.Error("New returns info for", enID, "expecting ", name, " found ", enName)
	case itName != name:
		t.Error("New returns info for", itID, "expecting ", name, " found ", itName)
	}
}
