package main

import (
	"reflect"
	"testing"
)

func TestSplitArgsFlagsBeforePositional(t *testing.T) {
	flags, pos := splitArgs(
		[]string{"-threshold", "90", "-json", "CONTEXT.yaml"},
		map[string]bool{"json": true},
	)
	wantFlags := []string{"-threshold", "90", "-json"}
	wantPos := []string{"CONTEXT.yaml"}
	if !reflect.DeepEqual(flags, wantFlags) {
		t.Errorf("flags = %v, want %v", flags, wantFlags)
	}
	if !reflect.DeepEqual(pos, wantPos) {
		t.Errorf("positional = %v, want %v", pos, wantPos)
	}
}

func TestSplitArgsFlagsAfterPositional(t *testing.T) {
	flags, pos := splitArgs(
		[]string{"xml", "-write"},
		map[string]bool{"write": true},
	)
	wantFlags := []string{"-write"}
	wantPos := []string{"xml"}
	if !reflect.DeepEqual(flags, wantFlags) {
		t.Errorf("flags = %v, want %v", flags, wantFlags)
	}
	if !reflect.DeepEqual(pos, wantPos) {
		t.Errorf("positional = %v, want %v", pos, wantPos)
	}
}

func TestSplitArgsFlagEqualsValue(t *testing.T) {
	flags, pos := splitArgs(
		[]string{"-file=custom.yaml", "json"},
		map[string]bool{},
	)
	wantFlags := []string{"-file=custom.yaml"}
	wantPos := []string{"json"}
	if !reflect.DeepEqual(flags, wantFlags) {
		t.Errorf("flags = %v, want %v", flags, wantFlags)
	}
	if !reflect.DeepEqual(pos, wantPos) {
		t.Errorf("positional = %v, want %v", pos, wantPos)
	}
}

func TestSplitArgsEmpty(t *testing.T) {
	flags, pos := splitArgs(nil, map[string]bool{})
	if flags != nil {
		t.Errorf("flags = %v, want nil", flags)
	}
	if pos != nil {
		t.Errorf("positional = %v, want nil", pos)
	}
}

func TestSplitArgsOnlyPositional(t *testing.T) {
	flags, pos := splitArgs(
		[]string{"xml"},
		map[string]bool{},
	)
	if flags != nil {
		t.Errorf("flags = %v, want nil", flags)
	}
	wantPos := []string{"xml"}
	if !reflect.DeepEqual(pos, wantPos) {
		t.Errorf("positional = %v, want %v", pos, wantPos)
	}
}

func TestSplitArgsMixedOrder(t *testing.T) {
	flags, pos := splitArgs(
		[]string{"-verbose", "xml", "-file", "custom.yaml"},
		map[string]bool{"verbose": true},
	)
	wantFlags := []string{"-verbose", "-file", "custom.yaml"}
	wantPos := []string{"xml"}
	if !reflect.DeepEqual(flags, wantFlags) {
		t.Errorf("flags = %v, want %v", flags, wantFlags)
	}
	if !reflect.DeepEqual(pos, wantPos) {
		t.Errorf("positional = %v, want %v", pos, wantPos)
	}
}
