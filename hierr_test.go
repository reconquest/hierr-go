package hierr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func ExampleError() {
	testcases := []error{
		Errorf(nil, ""),
		Errorf(nil, "simple error"),
		Errorf(nil, "integer: %d", 1),
		Errorf(errors.New("nested"), "top level"),
		Errorf(errors.New("nested"), "top level: %s", "formatting"),
		Errorf(Errorf(errors.New("low level"), "nested"), "top level"),

		Errorf(Errorf("string error", "nested"), "top level"),
		Errorf([]byte("byte"), "top level"),
	}

	for _, test := range testcases {
		fmt.Println()
		fmt.Println("{{{")
		fmt.Println(test.Error())
		fmt.Println("}}}")
	}

	fmt.Println()

	exiter = func(code int) {
		fmt.Println("exit code:", code)
	}

	tempfile, err := ioutil.TempFile(os.TempDir(), "stderr")
	if err != nil {
		panic(err)
	}

	os.Stderr = tempfile

	Fatalf("wow", "critical error")

	tempfile.Seek(0, 0)
	text, err := ioutil.ReadAll(tempfile)
	if err != nil {
		panic(err)
	}

	fmt.Println("stderr:\n" + string(text))

	// Output:
	//
	// {{{
	//
	// }}}
	//
	// {{{
	// simple error
	// }}}
	//
	// {{{
	// integer: 1
	// }}}
	//
	// {{{
	// top level
	// └─ nested
	// }}}
	//
	// {{{
	// top level: formatting
	// └─ nested
	// }}}
	//
	// {{{
	// top level
	// └─ nested
	//    └─ low level
	// }}}
	//
	// {{{
	// top level
	// └─ nested
	//    └─ string error
	// }}}
	//
	// {{{
	// top level
	// └─ byte
	// }}}
	//
	// exit code: 1
	// stderr:
	// critical error
	// └─ wow
}

func ExampleBranchDelimiter() {
	defer func() {
		BranchDelimiter = BranchDelimiterBox
	}()

	BranchDelimiter = "* "

	testcases := []error{
		Errorf(Errorf(errors.New("third"), "second"), "top level"),
	}

	for _, test := range testcases {
		fmt.Println()
		fmt.Println("{{{")
		fmt.Println(test.Error())
		fmt.Println("}}}")
	}

	// Output:
	//
	// {{{
	// top level
	// * second
	//    * third
	// }}}
}

func ExampleBranchIndent() {
	defer func() {
		BranchIndent = 3
	}()

	BranchIndent = 0

	testcases := []error{
		Errorf(Errorf(errors.New("third"), "second"), "top level"),
	}

	for _, test := range testcases {
		fmt.Println()
		fmt.Println("{{{")
		fmt.Println(test.Error())
		fmt.Println("}}}")
	}

	// Output:
	//
	// {{{
	// top level
	// └─ second
	// └─ third
	// }}}
}

func ExamplePush() {
	testcases := []error{
		Push(
			"the godfather",
			Push(
				"son A",
				"A's son 1",
				Push(
					"A's son 2",
					Push("2' son X",
						Push("X's son @"),
						Push("X's son #"),
					),
				),
			),
			Push("son B",
				errors.New("B's son 1"),
				errors.New("B's son 2"),
				Push("orphan"),
			),
			Errorf(
				"B's son 1",
				"son B",
			),
			errors.New("police"),
		),
	}

	for _, test := range testcases {
		fmt.Println()
		fmt.Println("{{{")
		fmt.Println(test.Error())
		fmt.Println("}}}")
	}

	// Output:
	//
	// {{{
	//the godfather
	//├─ son A
	//│  ├─ A's son 1
	//│  │
	//│  └─ A's son 2
	//│     └─ 2' son X
	//│        ├─ X's son @
	//│        └─ X's son #
	//│
	//├─ son B
	//│  ├─ B's son 1
	//│  ├─ B's son 2
	//│  └─ orphan
	//│
	//├─ son B
	//│  └─ B's son 1
	//│
	//└─ police
	// }}}
}

type smartError struct {
	Text string
	Err  error
}

func (smart smartError) HierarchicalError() string {
	return Errorf(smart.Err, smart.Text).Error()
}

func ExampleHierarchicalError() {
	testcases := []error{
		Errorf(
			Errorf(
				smartError{"smart", errors.New("hierarchical")},
				"second",
			),
			"top level",
		),
		Errorf(
			Errorf(
				fmt.Sprintf(
					"%s",
					smartError{"smart plain", errors.New("error")},
				),
				"second",
			),
			"top level",
		),
		Push(
			smartError{"smart", errors.New("hierarchical")},
			smartError{"smart", errors.New("hierarchical")},
		),
		Push(
			smartError{"smart", errors.New("hierarchical")},
			Push(
				smartError{"smart", errors.New("hierarchical")},
				smartError{"smart", errors.New("hierarchical")},
				smartError{"smart", Errorf("nested", "top")},
			),
		),
	}

	for _, test := range testcases {
		fmt.Println()
		fmt.Println("{{{")
		fmt.Println(test.Error())
		fmt.Println("}}}")
	}

	// Output:
	//
	// {{{
	// top level
	// └─ second
	//    └─ smart
	//       └─ hierarchical
	// }}}
	//
	// {{{
	// top level
	// └─ second
	//    └─ {smart plain error}
	// }}}
	//
	// {{{
	// smart
	// └─ hierarchical
	// └─ smart
	//    └─ hierarchical
	// }}}
	//
	// {{{
	// smart
	// └─ hierarchical
	// └─ smart
	//    └─ hierarchical
	//    ├─ smart
	//    │  └─ hierarchical
	//    └─ smart
	//       └─ top
	//          └─ nested
	// }}}
}
