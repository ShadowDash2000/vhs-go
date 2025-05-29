package errorcollector

import "errors"

type ErrorCollector struct {
	errs []error
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{}
}

func (ec *ErrorCollector) Collect(fn func() error) {
	if err := fn(); err != nil {
		ec.errs = append(ec.errs, err)
	}
}

func (ec *ErrorCollector) Error() error {
	if len(ec.errs) == 0 {
		return nil
	}
	return errors.Join(ec.errs...)
}

func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errs) > 0
}
