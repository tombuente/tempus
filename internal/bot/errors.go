package bot

import (
	"errors"
	"fmt"
)

var ErrSQLInternal = errors.New("sql database internal error")

func joinErrors(err1 error, err2 error) error {
	return fmt.Errorf("%w: %w", err1, err2)
}
