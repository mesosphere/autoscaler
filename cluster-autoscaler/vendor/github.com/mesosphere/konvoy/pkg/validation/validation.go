package validation

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// To ensure that valid cluster names are used and elbs can be created
var validHostnameRegex = regexp.MustCompile(`^((([a-z0-9]|[a-z0-9][-_.]{1})*[a-z0-9]))$`)

func ValidateClusterName(clusterName string, fldPath *field.Path) []Error {
	var allErrs []Error

	if clusterName == "" {
		allErrs = append(allErrs, WrapFieldError(field.Required(fldPath, "cannot have empty cluster name")))
	} else if !validateDomainName(clusterName) {
		allErrs = append(allErrs, WrapFieldError(field.Invalid(fldPath, clusterName, "invalid cluster name, must contain only 'a-z, 0-9, . - and _'")))
	}

	return allErrs
}

func validateDomainName(domain string) bool {
	return validHostnameRegex.MatchString(domain)
}

type Errors []Error

func (errors Errors) Unwrap() field.ErrorList {
	errorList := make(field.ErrorList, 0, len(errors))
	for _, e := range errors {
		errorList = append(errorList, e.Err)
	}
	return errorList
}

type Error struct {
	Action string
	Change string
	Reason string

	Err *field.Error
}

func (e *Error) Error() string {
	errorMsgs := make([]string, 0, 4)
	if e.Action != "" {
		errorMsgs = append(errorMsgs, fmt.Sprintf("action required: %s", e.Action))
	}
	if e.Change != "" {
		errorMsgs = append(errorMsgs, fmt.Sprintf("change: %s", e.Change))
	}
	if e.Reason != "" {
		errorMsgs = append(errorMsgs, fmt.Sprintf("reason: %s", e.Reason))
	}
	if e.Err != nil {
		errorMsgs = append(errorMsgs, fmt.Sprintf("error: %v", e.Err))
	}
	return strings.Join(errorMsgs, ": ")
}

func WrapFieldError(err *field.Error) Error {
	return Error{Err: err}
}
