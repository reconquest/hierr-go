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
	"os"
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
	Nested interface{}
}

var (
	exiter = os.Exit
)

// Error is either `error` or string.
type NestedError interface{}

// Errorf creates new hierarchy error.
//
// Have same semantics as `fmt.Errorf()`.
//
// With nestedError == nil call will be equal to `fmt.Errorf()`.
func Errorf(
	nestedError NestedError,
	message string,
	args ...interface{},
) error {
	return Error{
		Message: fmt.Sprintf(message, args...),
		Nested:  nestedError,
	}
}

func Fatalf(
	nestedError NestedError,
	message string,
	args ...interface{},
) {
	fmt.Println(Errorf(nestedError, message, args...))
	exiter(1)
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
			fmt.Sprintf("%s", err.Nested),
			"\n",
			"\n"+strings.Repeat(" ", BranchIndent),
			-1,
		)
}
