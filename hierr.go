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
package hierr // import "github.com/reconquest/hierr-go"

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	// BranchDelimiterASCII represents a simple ASCII delimiter for hierarchy
	// branches.
	//
	// Use: hierr.BranchDelimiter = hierr.BranchDelimiterASCII
	BranchDelimiterASCII = `\_ `

	// BranchDelimiterBox represents UTF8 delimiter for hierarchy branches.
	//
	// Use: hierr.BranchDelimiter = hierr.BranchDelimiterBox
	BranchDelimiterBox = `└─ `

	// BranchChainerASCII represents a simple ASCII chainer for hierarchy
	// branches.
	//
	// Use: hierr.BranchChainer = hierr.BranchChainerASCII
	BranchChainerASCII = `| `

	// BranchChainerBox represents UTF8 chainer for hierarchy branches.
	//
	// Use: hierr.BranchChainer = hierr.BranchChainerBox
	BranchChainerBox = `│ `

	// BranchSplitterASCII represents a simple ASCII splitter for hierarchy
	// branches.
	//
	// Use: hierr.BranchSplitter = hierr.BranchSplitterASCII
	BranchSplitterASCII = `+ `

	// BranchSplitterBox represents UTF8 splitter for hierarchy branches.
	//
	// Use: hierr.BranchSplitter = hierr.BranchSplitterBox
	BranchSplitterBox = `├─ `
)

var (
	// BranchDelimiter set delimiter each nested error text will be started
	// from.
	BranchDelimiter = BranchDelimiterBox

	// BranchChainer set chainer each nested error tree text will be started
	// from.
	BranchChainer = BranchChainerBox

	// BranchSplitter set splitter each nested errors splitted by.
	BranchSplitter = BranchSplitterBox

	// BranchIndent set number of spaces each nested error will be indented by.
	BranchIndent = 3
)

// Error represents hierarchy error, linked with nested error.
type Error struct {
	// Reason error, which can be hierr.Error as well.
	Reason Reason

	// Message is formatter error message, which will be reported when Error()
	// will be invoked.
	Message string

	// Context is a key-pair linked list, which represents runtime context
	// of the error.
	Context *ErrorContext
}

// HierarchicalError represents interface, which methods will be used instead
// of calling String() and Error() methods.
type HierarchicalError interface {
	// Error returns hierarchical string representation.
	Error() string

	// GetReasons returns slice of nested errors.
	GetReasons() []Reason

	// GetMessage returns top-level error message.
	GetMessage() string
}

var (
	exiter = os.Exit
)

// Reason is either `error` or string.
type Reason interface{}

// Errorf creates new hierarchy error.
//
// With reason == nil call will be equal to `fmt.Errorf()`.
func Errorf(
	reason Reason,
	message string,
	args ...interface{},
) error {
	return Error{
		Message: fmt.Sprintf(message, args...),
		Reason:  reason,
	}
}

// Fatalf creates new hierarchy error, prints to stderr and exit 1
//
// Have same semantics as `hierr.Errorf()`.
func Fatalf(
	reason Reason,
	message string,
	args ...interface{},
) {
	fmt.Fprintln(os.Stderr, Errorf(reason, message, args...))
	exiter(1)
}

// Error returns string representation of hierarchical error. If no nested
// error was specified, then only current error message will be returned.
func (err Error) Error() string {
	err.Context.Walk(func(name string, value interface{}) {
		err = Push(err, Push(fmt.Sprintf("%s: %s", name, value))).(Error)
	})

	switch value := err.Reason.(type) {
	case nil:
		return err.Message

	case []Reason:
		return formatReasons(err, value)

	default:
		return err.Message + "\n" +
			BranchDelimiter +
			strings.Replace(
				String(err.Reason),
				"\n",
				"\n"+strings.Repeat(" ", BranchIndent),
				-1,
			)
	}
}

// GetReasons returns nested errors, embedded into error.
func (err Error) GetReasons() []Reason {
	if err.Reason == nil {
		return nil
	}

	if reasons, ok := err.Reason.([]Reason); ok {
		return reasons
	} else {
		return []Reason{err.Reason}
	}
}

// GetMessage returns error message
func (err Error) GetMessage() string {
	if err.Message == "" {
		return fmt.Sprint(err.Reason)
	} else {
		return err.Message
	}
}

// GetContext returns context
func (err Error) GetContext() *ErrorContext {
	return err.Context
}

// Descend calls specified callback for every nested hierarchical error.
func (err Error) Descend(callback func(Error)) {
	for _, reason := range err.GetReasons() {
		if reason, ok := reason.(Error); ok {
			callback(reason)

			reason.Descend(callback)
		}
	}
}

// Push creates new hierarchy error with multiple branches separated by
// separator, delimited by delimiter and prolongated by prolongator.
func Push(reason Reason, reasons ...Reason) error {
	parent, ok := reason.(Error)
	if !ok {
		parent = Error{
			Message: String(reason),
		}
	}

	return Error{
		Message: parent.Message,
		Reason:  append(parent.GetReasons(), reasons...),
	}
}

// Context creates new context list, which can be used to produce context-rich
// hierarchical error.
func Context(key string, value interface{}) *ErrorContext {
	return &ErrorContext{
		Key:   key,
		Value: value,
	}
}

// String returns hierarchy-aware string representation of passed object.
func String(object interface{}) string {
	if hierr, ok := object.(HierarchicalError); ok {
		return hierr.Error()
	}

	if err, ok := object.(error); ok {
		return err.Error()
	}

	return fmt.Sprintf("%s", object)
}

func formatReasons(err Error, reasons []Reason) string {
	message := err.Message

	prolongate := false
	for _, reason := range reasons {
		if reasons, ok := reason.(HierarchicalError); ok {
			if len(reasons.GetReasons()) > 0 {
				prolongate = true
				break
			}
		}
	}

	for index, reason := range reasons {
		var (
			splitter      = BranchSplitter
			chainer       = BranchChainer
			chainerLength = len([]rune(BranchChainer))
		)

		if index == len(reasons)-1 {
			splitter = BranchDelimiter
			chainer = strings.Repeat(" ", chainerLength)
		}

		indentation := chainer
		if BranchIndent >= chainerLength {
			indentation += strings.Repeat(" ", BranchIndent-chainerLength)
		}

		prolongator := ""
		if prolongate && index < len(reasons)-1 {
			prolongator = "\n" + strings.TrimRightFunc(
				chainer, unicode.IsSpace,
			)
		}

		if message != "" {
			message = message + "\n" + splitter
		}

		message += strings.Replace(
			String(reason),
			"\n",
			"\n"+indentation,
			-1,
		)
		message += prolongator
	}

	return message
}
