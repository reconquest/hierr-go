// Package hierr provides a simple way to return and display hierarchical
// errors.
//
// Transforms:
//
//         can't pull remote 'origin': can't run git fetch 'origin' 'refs/tokens/*:refs/tokens/*': exit status 128
//
// Into:
//
//         can't pull remote 'origin'
//         └─ can't run git fetch 'origin' 'refs/tokens/*:refs/tokens/*'
//            └─ exit status 128
package hierr

import (
	"fmt"
	"strings"
)

const (
	// BranchDelimiterASCII represents a simple ASCII delimiter for hierarcy
	// branches.
	//
	// Use: hierr.BranchDelimiter = hierr.BranchDelimiterASCII
	BranchDelimiterASCII = `\_ `

	// BranchDelimiterBox represents UTF8 delimiter for hierarcy branches.
	//
	// Use: hierr.BranchDelimiter = hierr.BranchDelimiterBox
	BranchDelimiterBox = `└─ `
)

var (
	// BranchDelimiter set delimiter each nested error text will be started
	// from.
	BranchDelimiter = BranchDelimiterBox

	// BranchIndent set number of spaces each nested error will be indented by.
	BranchIndent = 3
)

// Error represents hierarchy error, linked with nested error.
type Error struct {
	// Message is formatter error message, which will be reported when Error()
	// will be invoked.
	Message string

	// Nested error, which can be hierr.Error as well.
	Nested error
}

// Errorf creates new hierr.Error.
//
// Have same semantics as `fmt.Errorf()`.
//
// With nestedError == nil method returns nil error.
func Errorf(nestedError error, message string, args ...interface{}) error {
	if nestedError == nil {
		return nil
	}

	return Error{
		Message: fmt.Sprintf(message, args...),
		Nested:  nestedError,
	}
}

// Error returns string representation of hierarchical error. If no nested
// error was specified, then only current error message will be returned.
func (err Error) Error() string {
	if err.Nested == nil {
		return err.Message
	}

	return err.Message + "\n" +
		BranchDelimiter +
		strings.Replace(
			err.Nested.Error(),
			"\n",
			"\n"+strings.Repeat(" ", BranchIndent),
			-1,
		)
}
