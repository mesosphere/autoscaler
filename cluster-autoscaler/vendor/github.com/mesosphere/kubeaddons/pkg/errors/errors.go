package errors

import "github.com/pkg/errors"

// Push is used to add an error to a Stack if it exists,
// or create one and add it if it doesn't.
func Push(stack **Stack, errs ...error) {
	if *stack == nil {
		*stack = newStack()
	}
	(*stack).push(errs...)
}

// Pop pops an error off the Stack. If it's the last one,
// delete the stack.
func Pop(stack **Stack) error {
	if *stack == nil {
		return nil
	}
	err := (*stack).pop()
	if (*stack).Len() == 0 {
		(*stack) = nil
	}
	return err
}

// NewStack provides a new Stack error
func newStack() *Stack {
	return &Stack{}
}

// Stack is a stack of errors that can be treated as a single wrapped error,
// or can be type casted to Stack to inspect the stack.
type Stack struct {
	Errors []error
}

// Push pushes an error onto the Stack
func (e *Stack) push(errs ...error) {
	for _, err := range errs {
		e.Errors = append(e.Errors, err)
	}
}

// Pop pops an error off the Stack
func (e *Stack) pop() error {
	err := e.Errors[len(e.Errors)-1]
	e.Errors = e.Errors[:len(e.Errors)-1]
	return err
}

// Error implements the error interface for Stack
func (e *Stack) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	var wrappedError error
	for i, err := range e.Errors {
		if i == 0 {
			wrappedError = err
		} else {
			wrappedError = errors.Wrap(wrappedError, err.Error())
		}
	}
	return wrappedError.Error()
}

// Len provides the size of the stack
func (e *Stack) Len() int {
	return len(e.Errors)
}
