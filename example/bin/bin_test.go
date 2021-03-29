package bin

import "testing"

func Test(t *testing.T) {
	res := createKFC()
	if res != "KFC" {
		t.Error("Error", res)
	}
}
