// Package secretstore provides an interface for secret store implementations.
package secretstore

type Service interface {
	GetSecretValue(secretName string) (string, error)
}
