package starlarkrepl

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type testCallableArgs struct {
	args []string
}

func (c testCallableArgs) String() string        { return "callable_kwargs" }
func (c testCallableArgs) Type() string          { return "callable_kwargs" }
func (c testCallableArgs) Freeze()               {}
func (c testCallableArgs) Truth() starlark.Bool  { return true }
func (c testCallableArgs) Hash() (uint32, error) { return 0, nil }
func (c testCallableArgs) Name() string          { return "callable_kwargs" }
func (c testCallableArgs) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}
func (c testCallableArgs) ArgNames() []string { return c.args }

func TestAutoComplete(t *testing.T) {
	mod := &starlarkstruct.Module{
		Name: "hello",
		Members: starlark.StringDict{
			"world": starlark.String("world"),
			"dict": func() *starlark.Dict {
				d := starlark.NewDict(2)
				d.SetKey(starlark.String("key"), starlark.String("value"))
				return d
			}(),
			"func": testCallableArgs{[]string{
				"kwarg",
			}},
		},
	}

	for _, tt := range []struct {
		name    string
		globals starlark.StringDict
		line    string
		want    []string
	}{{
		name: "simple",
		globals: map[string]starlark.Value{
			"abc": starlark.String("hello"),
		},
		line: "a",
		want: []string{"abc", "all", "any"},
	}, {
		name: "simple_semi",
		globals: map[string]starlark.Value{
			"abc": starlark.String("hello"),
		},
		line: "abc = \"hello\"; a",
		want: []string{
			"abc = \"hello\"; abc",
			"abc = \"hello\"; all",
			"abc = \"hello\"; any",
		},
	}, {
		name: "assignment",
		globals: map[string]starlark.Value{
			"abc": starlark.String("hello"),
		},
		line: "abc = a",
		want: []string{
			"abc = abc",
			"abc = all",
			"abc = any",
		},
	}, {
		name: "nest",
		globals: map[string]starlark.Value{
			"hello": mod,
		},
		line: "hello.wo",
		want: []string{"hello.world"},
	}, {
		name: "dict",
		globals: map[string]starlark.Value{
			"abc":   starlark.String("hello"),
			"hello": mod,
		},
		line: "hello.dict[ab",
		want: []string{"hello.dict[abc"},
	}, {
		name: "dict_string",
		globals: map[string]starlark.Value{
			"hello": mod,
		},
		line: "hello.dict[\"",
		want: []string{"hello.dict[\"key\"]"},
	}, {
		name: "call",
		globals: map[string]starlark.Value{
			"func": testCallableArgs{[]string{
				"arg_one", "arg_two",
			}},
		},
		line: "func(arg_",
		want: []string{"func(arg_one = ", "func(arg_two = "},
	}, {
		name: "call_multi",
		globals: map[string]starlark.Value{
			"func": testCallableArgs{[]string{
				"arg_one", "arg_two",
			}},
		},
		line: "func(arg_one = func(), arg_",
		want: []string{
			"func(arg_one = func(), arg_one = ",
			"func(arg_one = func(), arg_two = ",
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			c := completer{tt.globals}
			got := c.complete(tt.line)

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("%v != %v", tt.want, got)
			}
		})
	}
}
