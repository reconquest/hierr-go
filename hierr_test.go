package hierr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorf_CanFormatEmptyError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(Errorf(nil, ""), "")
}

func TestErrorf_CanFormatSimpleStringError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(Errorf(nil, "simple error"), "simple error")
}

func TestErrorf_CanFormatSimpleStringErrorWithArgs(t *testing.T) {
	test := assert.New(t)

	test.EqualError(Errorf(nil, "integer: %d", 9), "integer: 9")
}

func TestErrorf_CanFormatErrorWithSimpleReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Errorf(errors.New("reason"), "everything has a reason"),
		output(
			"everything has a reason",
			"└─ reason",
		),
	)
}

func TestErrorf_CanFormatErrorWithSimpleReasonAndArgs(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Errorf(errors.New("reason"), "reasons: %d", 1),
		output(
			"reasons: 1",
			"└─ reason",
		),
	)
}

func TestErrorf_CanFormatHierarchicalReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Errorf(Errorf(errors.New("reason"), "cause"), "karma"),
		output(
			"karma",
			"└─ cause",
			"   └─ reason",
		),
	)
}

func TestErrorf_CanFormatHierarchicalReasonWithSimpleReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Errorf(Errorf("reason", "cause"), "karma"),
		output(
			"karma",
			"└─ cause",
			"   └─ reason",
		),
	)
}

func TestErrorf_CanFormatAnyReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Errorf([]byte("self"), "no"),
		output(
			"no",
			"└─ self",
		),
	)
}

func TestFatalf_CanUseCustomExiter(t *testing.T) {
	test := assert.New(t)

	defer func() {
		exiter = os.Exit
	}()

	code := 0
	exiter = func(status int) {
		code = status
	}

	stderr := os.Stderr
	defer func() {
		os.Stderr = stderr
	}()

	var err error

	os.Stderr, err = ioutil.TempFile(os.TempDir(), "stderr")
	test.NoError(err)

	Fatalf("wow", "critical error")

	_, err = os.Stderr.Seek(0, 0)
	test.NoError(err)

	message, err := ioutil.ReadAll(os.Stderr)
	test.NoError(err)
	test.Equal(1, code)
	test.Equal(
		output(
			"critical error",
			"└─ wow\n",
		),
		string(message),
	)
}

func TestCanSetBranchDelimiter(t *testing.T) {
	test := assert.New(t)

	delimiter := BranchDelimiter
	defer func() {
		BranchDelimiter = delimiter
	}()

	BranchDelimiter = "* "

	test.EqualError(
		Errorf(Errorf("first", "second"), "third"),
		output(
			"third",
			"* second",
			"   * first",
		),
	)
}

func TestCanSetBranchIndent(t *testing.T) {
	test := assert.New(t)

	indent := BranchIndent
	defer func() {
		BranchIndent = indent
	}()

	BranchIndent = 0

	test.EqualError(
		Errorf(Errorf("first", "second"), "third"),
		output(
			"third",
			"└─ second",
			"└─ first",
		),
	)
}

func TestContext_CanAddMultipleKeyValues(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Context("host", "example.com").Context("operation", "resolv").Errorf(
			"system error",
			"unable to resolve",
		),
		output(
			"unable to resolve",
			"├─ system error",
			"├─ host: example.com",
			"└─ operation: resolv",
		),
	)
}

func TestContext_CanAddWithoutHierarchy(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Context("host", "example.com").Reason(
			Context("operation", "resolv").Reason(
				"system error",
			),
		),
		output(
			"system error",
			"├─ operation: resolv",
			"└─ host: example.com",
		),
	)
}

func TestContext_CanAddToRootError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Context("host", "example.com").Errorf(
			"system error",
			"unable to resolve",
		),
		output(
			"unable to resolve",
			"├─ system error",
			"└─ host: example.com",
		),
	)
}

func TestContext_CanAddToReasonError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Context("host", "example.com").Errorf(
			Context("os", "linux").Reason(
				"system error",
			),
			"unable to resolve",
		),
		output(
			"unable to resolve",
			"├─ system error",
			"│  └─ os: linux",
			"│",
			"└─ host: example.com",
		),
	)
}

type customError struct {
	Text   string
	Reason error
}

func (err customError) Error() string {
	return Errorf(err.Reason, err.GetMessage()).Error()
}

func (err customError) GetNested() []Reason {
	return []Reason{err.Reason}
}

func (err customError) GetMessage() string {
	return strings.ToUpper(err.Text)
}

func TestCustomHierarchicalError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Errorf(
			customError{"upper", errors.New("hierarchical")},
			"example of custom error",
		),
		output(
			"example of custom error",
			"└─ UPPER",
			"   └─ hierarchical",
		),
	)
}

func ExampleContext_MultipleKeyValues() {
	foo := func(arg string) error {
		return fmt.Errorf("unable to foo on %s", arg)
	}

	bar := func() error {
		err := foo("zen")
		if err != nil {
			return Context("method", "foo").Context("arg", "zen").Reason(err)
		}

		return nil
	}

	err := bar()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	//
	// unable to foo on zen
	// ├─ method: foo
	// └─ arg: zen
}

func ExampleContext_NestedErrors() {
	foo := func(arg string) error {
		return fmt.Errorf("unable to foo on %s", arg)
	}

	bar := func() error {
		err := foo("zen")
		if err != nil {
			return Context("arg", "zen").Reason(err)
		}

		return nil
	}

	baz := func() error {
		err := bar()
		if err != nil {
			return Context("operation", "foo").Errorf(
				err,
				"unable to perform critical operation",
			)
		}

		return nil
	}

	err := baz()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	//
	// unable to perform critical operation
	// ├─ unable to foo on zen
	// │  └─ arg: zen
	// │
	// └─ operation: foo
}

func ExampleContext_AddNestedContext() {
	foo := func() error {
		return fmt.Errorf("unable to foo")
	}

	bar := func() error {
		err := foo()
		if err != nil {
			return Context("level", "bar").Reason(err)
		}

		return nil
	}

	baz := func() error {
		err := bar()
		if err != nil {
			return Context("level", "baz").Reason(err)
		}

		return nil
	}

	err := baz()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	//
	// unable to foo
	// ├─ level: bar
	// └─ level: baz
}

func ExampleContext_UseCustomLoggingFormat() {
	// solve function represents deepest function in the call stack
	solve := func(koan string) error {
		return fmt.Errorf("no solution available for %q", koan)
	}

	// think represents function, which calls solve function
	think := func() error {
		err := solve("what was your face before your parents were born?")
		if err != nil {
			return Context("though", "koan").Reason(err)
		}

		return nil
	}

	// realize represents top-level function, which calls think function
	realize := func() error {
		context := Context("doing", "realization")
		err := think()
		if err != nil {
			return context.Context("action", "thinking").Errorf(
				err,
				"unable to attain realization",
			)
		}

		return nil
	}

	// log represents custom logging function, which writes structured logs,
	// like logrus in format [LEVEL] message: key1=value1 key2=value2
	log := func(level string, message string, kv ...interface{}) {
		fmt.Printf("[%s] %s:", level, message)

		for i := 0; i < len(kv); i += 2 {
			fmt.Printf(" %s=%q", kv[i], kv[i+1])
		}

		fmt.Println()
	}

	err := realize()
	if err != nil {
		if err, ok := err.(Error); ok {
			// following call will write all nested errors
			err.Descend(func(err Error) {
				log(
					"ERROR",
					err.GetMessage(),
					err.GetContext().GetKeyValuePairs()...,
				)
			})

			// this call will write only root-level error
			log(
				"FATAL",
				err.GetMessage(),
				err.GetContext().GetKeyValuePairs()...,
			)
		}
	}

	// Output:
	//
	// [ERROR] no solution available for "what was your face before your parents were born?": though="koan"
	// [FATAL] unable to attain realization: doing="realization" action="thinking"
}

func output(lines ...string) string {
	return strings.Join(lines, "\n")
}
