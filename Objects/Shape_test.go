package Objects

import (
	"reflect"
	"testing"
)

func TestShapeRotate90(t *testing.T) {
	in := Shape{Blocks: [][]string{
		{"X", "X", "X"},
		{" ", "X", " "},
	}}
	want := Shape{Blocks: [][]string{
		{" ", "X"},
		{"X", "X"},
		{" ", "X"},
	}}
	got := in.Rotate()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Rotate = %v, want %v", got, want)
	}
}

func TestShapeRotateFourTimesIsIdentity(t *testing.T) {
	for name, shp := range map[string]Shape{
		"Shape1": Shape1, "Shape2": Shape2, "Shape3": Shape3, "Shape4": Shape4,
		"Shape5": Shape5, "Shape6": Shape6, "Shape7": Shape7, "Shape9": Shape9,
	} {
		out := shp.Rotate().Rotate().Rotate().Rotate()
		if !reflect.DeepEqual(out, shp) {
			t.Errorf("%s: 4 rotations should be identity, got %v", name, out)
		}
	}
}

func TestShapeRotateSingleCell(t *testing.T) {
	got := Shape8.Rotate()
	if !reflect.DeepEqual(got, Shape8) {
		t.Errorf("Shape8 rotated = %v, want %v", got, Shape8)
	}
}

func TestShapeRotateEmpty(t *testing.T) {
	in := Shape{}
	got := in.Rotate()
	if !reflect.DeepEqual(got, in) {
		t.Errorf("empty shape rotated should be empty, got %v", got)
	}
}
