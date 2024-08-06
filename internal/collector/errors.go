package collector

import "errors"

var (
	ErrCommandNotFound    = errors.New("command not found")
	ErrCommandExec        = errors.New("command execution")
	ErrInvalidParseResult = errors.New("invalid parseing result")
)
