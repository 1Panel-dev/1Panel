package iptables_helper

import "testing"

func TestHasBaseChainBinding(t *testing.T) {
	if hasBaseChainBinding("-P INPUT ACCEPT") {
		t.Fatal("input policy was treated as a 1Panel binding")
	}
	if !hasBaseChainBinding("-A INPUT -j " + BasicAfterChain) {
		t.Fatal("partial iptables binding was not detected")
	}
}
