// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worksheets

import (
	"fmt"
)

type expression interface {
	Args() []string
	Compute(ws *Worksheet) (Value, error)
}

// Assert that all expressions implement the expression interface
var _ = []expression{
	&Undefined{},
	&Number{},
	&Text{},
	&Bool{},

	&tExternal{},
	&ePlugin{},
	&tVar{},
	&tUnop{},
	&tBinop{},
	&tReturn{},
}

func (e *tExternal) Args() []string {
	panic(fmt.Sprintf("unresolved plugin in worksheet"))
}

func (e *tExternal) Compute(ws *Worksheet) (Value, error) {
	panic(fmt.Sprintf("unresolved plugin in worksheet(%s)", ws.def.name))
}

func (e *Undefined) Args() []string {
	return nil
}

func (e *Undefined) Compute(ws *Worksheet) (Value, error) {
	return e, nil
}

func (e *Number) Args() []string {
	return nil
}

func (e *Number) Compute(ws *Worksheet) (Value, error) {
	return e, nil
}

func (e *Text) Args() []string {
	return nil
}

func (e *Text) Compute(ws *Worksheet) (Value, error) {
	return e, nil
}

func (e *Bool) Args() []string {
	return nil
}

func (e *Bool) Compute(ws *Worksheet) (Value, error) {
	return e, nil
}

func (e *tVar) Args() []string {
	return []string{e.name}
}

func (e *tVar) Compute(ws *Worksheet) (Value, error) {
	return ws.Get(e.name)
}

func (e *tUnop) Args() []string {
	return e.expr.Args()
}

func (e *tUnop) Compute(ws *Worksheet) (Value, error) {
	result, err := e.expr.Compute(ws)
	if err != nil {
		return nil, err
	}

	if _, ok := result.(*Undefined); ok {
		return result, nil
	}

	switch e.op {
	case opNot:
		bResult, ok := result.(*Bool)
		if !ok {
			return nil, fmt.Errorf("! on non-bool")
		}
		return &Bool{!bResult.value}, nil
	default:
		panic(fmt.Sprintf("not implemented for %s", e.op))
	}
}

func (e *tBinop) Args() []string {
	left := e.left.Args()
	right := e.right.Args()
	return append(left, right...)
}

func (e *tBinop) Compute(ws *Worksheet) (Value, error) {
	left, err := e.left.Compute(ws)
	if err != nil {
		return nil, err
	}

	// bool operations
	if e.op == opAnd || e.op == opOr {
		if _, ok := left.(*Undefined); ok {
			return left, nil
		}

		bLeft, ok := left.(*Bool)
		if !ok {
			return nil, fmt.Errorf("op on non-bool")
		}

		if (e.op == opAnd && !bLeft.value) || (e.op == opOr && bLeft.value) {
			return bLeft, nil
		}

		right, err := e.right.Compute(ws)
		if err != nil {
			return nil, err
		}

		if _, ok := right.(*Undefined); ok {
			return right, nil
		}

		bRight, ok := right.(*Bool)
		if !ok {
			return nil, fmt.Errorf("op on non-bool")
		}

		return bRight, nil
	}

	right, err := e.right.Compute(ws)
	if err != nil {
		return nil, err
	}

	// equality
	if e.op == opEqual {
		return &Bool{left.Equal(right)}, nil
	}
	if e.op == opNotEqual {
		return &Bool{!left.Equal(right)}, nil
	}

	// numerical operations
	nLeft, ok := left.(*Number)
	if !ok {
		return nil, fmt.Errorf("op on non-number")
	}

	if _, ok := left.(*Undefined); ok {
		return left, nil
	}

	nRight, ok := right.(*Number)
	if !ok {
		return nil, fmt.Errorf("op on non-number")
	}

	if _, ok := right.(*Undefined); ok {
		return right, nil
	}

	var result *Number
	switch e.op {
	case opPlus:
		result = nLeft.Plus(nRight)
	case opMinus:
		result = nLeft.Minus(nRight)
	case opMult:
		result = nLeft.Mult(nRight)
	case opDiv:
		if e.round == nil {
			return nil, fmt.Errorf("division without rounding mode")
		}
		return nLeft.Div(nRight, e.round.mode, e.round.scale), nil
	default:
		panic(fmt.Sprintf("not implemented for %s", e.op))
	}

	if e.round != nil {
		result = result.Round(e.round.mode, e.round.scale)
	}

	return result, nil
}

func (e *tReturn) Args() []string {
	return e.expr.Args()
}

func (e *tReturn) Compute(ws *Worksheet) (Value, error) {
	return e.expr.Compute(ws)
}

type ePlugin struct {
	computedBy ComputedBy
}

func (e *ePlugin) Args() []string {
	return e.computedBy.Args()
}

func (e *ePlugin) Compute(ws *Worksheet) (Value, error) {
	args := e.computedBy.Args()
	values := make([]Value, len(args), len(args))
	for i, arg := range args {
		value := ws.MustGet(arg)
		values[i] = value
	}
	return e.computedBy.Compute(values...), nil
}
