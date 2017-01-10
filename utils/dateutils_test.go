package utils

import "testing"

func TestTimestamp(t *testing.T) {
	t10 := Timestamp10()
	t13 := Timestamp13()

	t.Log("t10 = ", t10)
	t.Log("t13 = ", t13)
}
