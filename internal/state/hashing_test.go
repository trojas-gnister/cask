package state

import "testing"

func TestHashString(t *testing.T) {
	h := HashString("hello")
	if len(h) != 64 {
		t.Errorf("expected 64 char hex digest, got %d chars", len(h))
	}
	// SHA256 of "hello" is well-known
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if h != expected {
		t.Errorf("expected %s, got %s", expected, h)
	}
}

func TestHashStringDeterministic(t *testing.T) {
	a := HashString("test value")
	b := HashString("test value")
	if a != b {
		t.Error("same input should produce same hash")
	}
}

func TestHashStringDifferentInputs(t *testing.T) {
	a := HashString("foo")
	b := HashString("bar")
	if a == b {
		t.Error("different inputs should produce different hashes")
	}
}

func TestHashData(t *testing.T) {
	data := map[string]any{
		"packages": []string{"vim", "git"},
		"enabled":  true,
	}
	h, err := HashData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h) != 64 {
		t.Errorf("expected 64 char hex digest, got %d", len(h))
	}
}

func TestHashDataDeterministic(t *testing.T) {
	data := map[string]any{
		"b": "second",
		"a": "first",
	}
	h1, _ := HashData(data)
	h2, _ := HashData(data)
	if h1 != h2 {
		t.Error("same data should produce same hash")
	}
}

func TestHashDataMapKeyOrder(t *testing.T) {
	// Go's json.Marshal sorts map keys, so order shouldn't matter
	a := map[string]string{"z": "1", "a": "2"}
	b := map[string]string{"a": "2", "z": "1"}
	ha, _ := HashData(a)
	hb, _ := HashData(b)
	if ha != hb {
		t.Error("map key order should not affect hash")
	}
}

func TestHashDataNestedStructs(t *testing.T) {
	data := map[string]any{
		"outer": map[string]any{
			"inner": []string{"a", "b"},
		},
	}
	h, err := HashData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h) != 64 {
		t.Error("nested structure should hash successfully")
	}
}
