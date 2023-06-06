package ezg

import "testing"

func Test_AutopreloadingFields(t *testing.T) {
	type ChldT4NoPreload struct {
		Foo string
	}
	type ChldT3 struct {
		Foo         string
		NoPreloads  []ChldT4NoPreload `ezg:"no-preload"`
		NoPreloads2 []ChldT4NoPreload `ezg:"nopreload"`
	}
	type ChldT2 struct {
		ChildrenT3 []*ChldT3
	}
	type ChldT1 struct {
		ChildrenT2 []ChldT2
	}
	type Parent struct {
		Children []*ChldT1
	}
	want := []string{
		"Children", "Children.ChildrenT2", "Children.ChildrenT2.ChildrenT3",
	}
	got := autoPreloads(&Parent{})
	if len(want) != len(got) {
		t.Fatalf("wanted %d results, but got %d results", len(want), len(got))
		return
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("wanted %s result, but got %s results at index %d", want[i], got[i], i)
			return
		}
	}
}
