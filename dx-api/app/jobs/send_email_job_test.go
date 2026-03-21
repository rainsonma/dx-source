package jobs

import (
	"testing"
)

func TestSendEmailJob_Signature(t *testing.T) {
	job := &SendEmailJob{}
	if got := job.Signature(); got != "send_email" {
		t.Errorf("Signature() = %q, want %q", got, "send_email")
	}
}

func TestSendEmailJob_Handle_InsufficientArgs(t *testing.T) {
	job := &SendEmailJob{}

	tests := []struct {
		name string
		args []any
	}{
		{"no args", nil},
		{"one arg", []any{"to@example.com"}},
		{"two args", []any{"to@example.com", "subject"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := job.Handle(tt.args...)
			if err == nil {
				t.Error("expected error for insufficient args, got nil")
			}
		})
	}
}

func TestSendEmailJob_Handle_InvalidArgTypes(t *testing.T) {
	job := &SendEmailJob{}

	tests := []struct {
		name string
		args []any
	}{
		{"to is not string", []any{123, "subject", "<p>body</p>"}},
		{"subject is not string", []any{"to@example.com", 456, "<p>body</p>"}},
		{"html is not string", []any{"to@example.com", "subject", 789}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := job.Handle(tt.args...)
			if err == nil {
				t.Error("expected error for invalid arg type, got nil")
			}
		})
	}
}
