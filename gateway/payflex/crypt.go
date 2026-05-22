package payflex

import "github.com/zatrano/gopostr/crypt"

// newEnrollmentRequestID MPI enrollment istek kimliği üretir.
func newEnrollmentRequestID() (string, error) {
	return crypt.RandomString(24)
}
