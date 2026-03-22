package jobs

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/queue"
)

type SendEmailJob struct {
}

// Signature returns the unique signature of the job.
func (j *SendEmailJob) Signature() string {
	return "send_email"
}

// Handle executes the job.
// args[0]: to (string), args[1]: subject (string), args[2]: html (string)
func (j *SendEmailJob) Handle(args ...any) error {
	if len(args) < 3 {
		return fmt.Errorf("send_email job requires 3 args (to, subject, html), got %d", len(args))
	}

	to, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("send_email job: to must be a string")
	}
	subject, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("send_email job: subject must be a string")
	}
	html, ok := args[2].(string)
	if !ok {
		return fmt.Errorf("send_email job: html must be a string")
	}

	if err := facades.Mail().
		To([]string{to}).
		Content(mail.Content{Html: html}).
		Subject(subject).
		Send(); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}

// DispatchSendEmail dispatches the send email job onto the queue.
func DispatchSendEmail(to, subject, html string) error {
	return facades.Queue().Job(&SendEmailJob{}, []queue.Arg{
		{Type: "string", Value: to},
		{Type: "string", Value: subject},
		{Type: "string", Value: html},
	}).Dispatch()
}
