package v1beta2

import (
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory"
)

type ClientBuilder = inventory.ClientBuilder[KesselInventoryServiceClient]

func NewClientBuilder(target string) *ClientBuilder {
	return inventory.NewClientBuilder[KesselInventoryServiceClient](target, NewKesselInventoryServiceClient)
}
