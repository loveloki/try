package script

import (
	"fmt"
	"io"
	"os"

	"github.com/loveloki/try/internal/selector"
)

// ExecuteSideEffect 执行 mkdir/delete/rename/ship，不写 cd 脚本到 stdout。
// SelectCD 与 nil 结果直接忽略。
func ExecuteSideEffect(result *selector.SelectionResult) error {
	return executeSideEffect(io.Discard, os.Stderr, result)
}

func executeSideEffect(stdout, stderr io.Writer, result *selector.SelectionResult) error {
	if result == nil {
		return nil
	}
	switch result.Type {
	case selector.SelectMkdir, selector.SelectDelete, selector.SelectRename, selector.SelectShip:
		return ExecuteTo(stdout, stderr, result)
	case selector.SelectCD:
		return nil
	default:
		return fmt.Errorf(msgs().ErrUnknownOp, result.Type)
	}
}
