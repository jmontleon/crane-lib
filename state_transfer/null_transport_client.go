package state_transfer

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *NullTransport) CreateClient(c client.Client, t Transfer) error {
	return nil
}
