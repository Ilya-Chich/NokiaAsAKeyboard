package mapkeys

import "testing"

func TestLookup(t *testing.T) {
	a, ok := Lookup('a')
	if !ok || a.Code != KeyA || a.Shift {
		t.Fatalf("a: %+v ok=%v", a, ok)
	}
	upper, ok := Lookup('A')
	if !ok || upper.Code != KeyA || !upper.Shift {
		t.Fatalf("A: %+v ok=%v", upper, ok)
	}
	bang, ok := Lookup('!')
	if !ok || bang.Code != Key1 || !bang.Shift {
		t.Fatalf("!: %+v ok=%v", bang, ok)
	}
	if _, ok := Lookup('č'); ok {
		t.Fatal("serbian letter must not map")
	}
	for _, r := range "!@#$%^&*()_+-=[]\\{}|;':\",.<>/?`~ " {
		if _, ok := Lookup(r); !ok {
			t.Fatalf("missing %q", r)
		}
	}
}
