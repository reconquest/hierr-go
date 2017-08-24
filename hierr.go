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
	// Message is formatter error message, which will be reported when Error()
	// will be invoked.
	Message string

	// Nested error, which can be hierr.Error as well.
	Nested interface{}
}

// HierarchicalError represents interface, which methods will be used instead
// of calling String() and Error() methods.
type HierarchicalError interface {
	// HierarchicalError returns hierarhical string representation.
	HierarchicalError() string

	// GetNested returns slice of nested errors.
	GetNested() []NestedError

	// GetMessage returns top-level error message.
	GetMessage() string
}

var (
	exiter = os.Exit
)

// NestedError is either `error` or string.
type NestedError interface{}

// Errorf creates new hierarchy error.
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

// Fatalf creates new hierarchy error, prints to stderr and exit 1
//
// Have same semantics as `hierr.Errorf()`.
func Fatalf(
	nestedError NestedError,
	message string,
	args ...interface{},
) {
	fmt.Fprintln(os.Stderr, Errorf(nestedError, message, args...))
	exiter(1)
}

// Error returns string representation of hierarchical error. If no nested
// error was specified, then only current error message will be returned.
func (err Error) Error() string {
	switch children := err.Nested.(type) {
	case nil:
		return err.Message

	case []NestedError:
		return formatNestedError(err, children)

	default:
		return err.Message + "\n" +
			BranchDelimiter +
			strings.Replace(
				String(err.Nested),
				"\n",
				"\n"+strings.Repeat(" ", BranchIndent),
				-1,
			)
	}
}

// GetNested returns nested errors, embedded into error.
func (err Error) GetNested() []NestedError {
	children, ok := err.Nested.([]NestedError)
	if !ok {
		children = []NestedError{}
		if err.Nested != nil {
			children = append(children, err.Nested)
		}
	}

	return children
}

// GetMessage returns top-level error message.
func (err Error) GetMessage() string {
	return err.Message
}

// HierarchicalError returns pretty hierarchical rendering.
func (err Error) HierarchicalError() string {
	return err.Error()
}

// Push creates new hierarchy error with multiple branches separated by
// separator, delimited by delimiter and prolongated by prolongator.
func Push(topError NestedError, childError ...NestedError) error {
	parent, ok := topError.(Error)
	if !ok {
		parent = Error{
			Message: String(topError),
		}
	}

	children := parent.GetNested()

	children = append(children, childError...)

	return Error{
		Message: parent.Message,
		Nested:  children,
	}
}

// Context adds context to specified top-level node.
//
// Context can be passed to rest of the call to add multiple labels to
// given error:
//	hierr.Context(
//		err,
//		hierr.Context(`mailer`, `localhost:25`),
//		hierr.Context(`config`, `/path/to/config.toml`),
//	)
func Context(node NestedError, description ...NestedError) error {
	return Push(node, description...)
}

func String(object interface{}) string {
	if hierr, ok := object.(HierarchicalError); ok {
		return hierr.HierarchicalError()
	}

	if err, ok := object.(error); ok {
		return err.Error()
	}

	return fmt.Sprintf("%s", object)
}

func formatNestedError(err Error, children []NestedError) string {
	message := err.Message

	prolongate := false
	for _, child := range children {
		if childError, ok := child.(HierarchicalError); ok {
			errs := childError.GetNested()
			if len(errs) > 0 {
				prolongate = true
				break
			}
		}
	}

	for index, child := range children {
		var (
			splitter      = BranchSplitter
			chainer       = BranchChainer
			chainerLength = len([]rune(BranchChainer))
		)

		if index == len(children)-1 {
			splitter = BranchDelimiter
			chainer = strings.Repeat(" ", chainerLength)
		}

		indentation := chainer
		if BranchIndent >= chainerLength {
			indentation += strings.Repeat(" ", BranchIndent-chainerLength)
		}

		prolongator := ""
		if prolongate && index < len(children)-1 {
			prolongator = "\n" + strings.TrimRightFunc(
				chainer, unicode.IsSpace,
			)
		}

		message = message + "\n" +
			splitter +
			strings.Replace(
				String(child),
				"\n",
				"\n"+indentation,
				-1,
			) +
			prolongator
	}

	return message
}
