package snowflake

import "testing"

func TestGetId(t *testing.T) {
	worker, err := NewWorker(1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(worker.GetId())
}
