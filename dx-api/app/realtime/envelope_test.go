package realtime

import (
	"encoding/json"
	"testing"
)

func TestEnvelope_MarshalOmitsEmptyFields(t *testing.T) {
	env := Envelope{Op: OpSubscribe, Topic: "user:alice", ID: "req_1"}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(b)
	want := `{"op":"subscribe","topic":"user:alice","id":"req_1"}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_MarshalEventWithData(t *testing.T) {
	env := Envelope{Op: OpEvent, Topic: "pk:abc", Type: "pk_player_action", Data: map[string]string{"user_id": "alice"}}
	b, _ := json.Marshal(env)
	var back map[string]any
	_ = json.Unmarshal(b, &back)
	if back["op"] != "event" {
		t.Errorf("wrong op: %v", back["op"])
	}
	if back["topic"] != "pk:abc" {
		t.Errorf("wrong topic: %v", back["topic"])
	}
	if back["type"] != "pk_player_action" {
		t.Errorf("wrong type: %v", back["type"])
	}
}

func TestEnvelope_AckWithOKTrue(t *testing.T) {
	ok := true
	env := Envelope{Op: OpAck, ID: "req_1", OK: &ok}
	b, _ := json.Marshal(env)
	got := string(b)
	want := `{"op":"ack","id":"req_1","ok":true}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_AckWithOKFalseAndCode(t *testing.T) {
	ok := false
	env := Envelope{Op: OpAck, ID: "req_1", OK: &ok, Code: 40300, Message: "forbidden"}
	b, _ := json.Marshal(env)
	got := string(b)
	want := `{"op":"ack","id":"req_1","ok":false,"code":40300,"message":"forbidden"}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_ErrorOp(t *testing.T) {
	env := Envelope{Op: OpError, Code: 40022, Message: "envelope missing op field"}
	b, _ := json.Marshal(env)
	got := string(b)
	want := `{"op":"error","code":40022,"message":"envelope missing op field"}`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestEnvelope_UnmarshalRoundtrip(t *testing.T) {
	original := Envelope{Op: OpSubscribe, Topic: "group:xyz", ID: "req_42"}
	b, _ := json.Marshal(original)
	var back Envelope
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Op != original.Op || back.Topic != original.Topic || back.ID != original.ID {
		t.Errorf("roundtrip mismatch: %+v vs %+v", back, original)
	}
}

func TestEvent_CarriesTypeAndData(t *testing.T) {
	e := Event{Type: "pk_player_complete", Data: map[string]int{"score": 100}}
	if e.Type != "pk_player_complete" {
		t.Errorf("wrong type: %s", e.Type)
	}
	if m, ok := e.Data.(map[string]int); !ok || m["score"] != 100 {
		t.Errorf("wrong data: %+v", e.Data)
	}
}
