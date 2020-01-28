package v1beta1

// AzureProviderOptions describes azure provider related options
type AzureProviderOptions struct {
	Location        *string           `json:"location,omitempty"`
	VNET            *VNET             `json:"vnet,omitempty"`
	AvailabilitySet *AvailabilitySet  `json:"availabilitySet,omitempty"`
	LoadBalancer    *LoadBalancer     `json:"loadbalancer,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
}

// VNET contains the virtual network information required if using an existing virtual network
type VNET struct {
	Name          *string `json:"name,omitempty"`
	ResourceGroup *string `json:"resourceGroup,omitempty"`
	RouteTable    *string `json:"routeTable,omitempty"`
}

// AvailabilitySet contains the availability_set information
type AvailabilitySet struct {
	FaultDomainCount  *int32 `json:"faultDomainCount,omitempty"`
	UpdateDomainCount *int32 `json:"updateDomainCount,omitempty"`
}

// LoadBalancer contains details for the kube-apiserver LoadBalancer
type LoadBalancer struct {
	Internal *bool `json:"internal,omitempty"`
}

// AzureMachineOpts is azure specific options for machine
type AzureMachineOpts struct {
	SubnetIDs []string `json:"subnetIDs,omitempty"`
}
