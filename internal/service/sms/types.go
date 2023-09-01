package sms

import "context"

type Service interface {
	Send(ctx context.Context, template string, args []NamedArg, numbers []string) error
}

type NamedArg struct {
	Value string
	Name  string
}
