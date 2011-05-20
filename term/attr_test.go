package term

import "testing"

func TestAttr(t *testing.T) {
    term,err := Mine()
    if err != nil {
        t.Fatalf("Unable to acquire terminal handle.")
    }

    attr,err := term.GetAttributes()
    if err != nil {
        t.Fatalf("Unable to get terminal attributes.")
    }

    oldInput := attr.Input
    attr.Input ^= IGNCR
    err = term.SetAttributes(attr, NOW)
    if err != nil {
        t.Fatalf("Unable to set terminal attributes.")
    }

    attr,err = term.GetAttributes()
    if err != nil {
        t.Fatalf("Unable to get terminal attributes.")
    }

    mask := oldInput ^ attr.Input
    if mask != IGNCR {
        t.Fatalf("Attribute manipulation fails consistency check.")
    }

    attr.Input = oldInput
    term.SetAttributes(attr, NOW)
}

