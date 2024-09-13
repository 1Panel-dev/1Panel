package huawei

import "testing"

func TestCreate(t *testing.T) {
	p, err := NewProvider(
		"",
		"",
		"",
		"",
		"",
	)
	if err != nil {
		t.Fatal(err)
	}

	err = p.CreateOrUpdateRecord("11.4.51.4")
	if err != nil {
		t.Fatal(err)
	}
}
