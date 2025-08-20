package v1beta2

import (
	genericBuilder "github.com/project-kessel/kessel-sdk-go/kessel/inventory/internal/builder"
)

type ClientBuilder = genericBuilder.ClientBuilder[KesselInventoryServiceClient]

func NewClientBuilder(target string) *ClientBuilder {
	return genericBuilder.NewClientBuilder[KesselInventoryServiceClient](target, NewKesselInventoryServiceClient)
}
