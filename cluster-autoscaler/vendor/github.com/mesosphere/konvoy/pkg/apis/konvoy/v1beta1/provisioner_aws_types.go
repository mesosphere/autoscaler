package v1beta1

const (
	// PrefixARNInstanceProfile is the prefix in ARNs for instanceProfile
	PrefixARNInstanceProfile = "instance-profile/"
	// ARNSeparator is a specific separator for ARNs
	ARNSeparator = ":"
)

// AWSProviderOptions describes AWS provider related options
type AWSProviderOptions struct {
	Region            *string           `json:"region,omitempty"`
	VPC               *VPC              `json:"vpc,omitempty"`
	AvailabilityZones []string          `json:"availabilityZones,omitempty"`
	ELB               *ELB              `json:"elb,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// VPC contains the vpc information required if using an existing vpc
type VPC struct {
	ID                    *string `json:"ID,omitempty"`
	RouteTableID          *string `json:"routeTableID,omitempty"`
	InternetGatewayID     *string `json:"internetGatewayID,omitempty"`
	EnableInternetGateway *bool   `json:"enableInternetGateway,omitempty"`
	EnableVPCEndpoints    *bool   `json:"enableVPCEndpoints,omitempty"`
}

// ELB contains details for the kube-apiserver ELB
type ELB struct {
	Internal  *bool    `json:"internal,omitempty"`
	SubnetIDs []string `json:"subnetIDs,omitempty"`
}

// AWSMachineOpts is aws specific options for machine
type AWSMachineOpts struct {
	IAM       *IAM     `json:"iam,omitempty"`
	SubnetIDs []string `json:"subnetIDs,omitempty"`
}

// IAM contains role information to use instead of creating one
type IAM struct {
	InstanceProfile *InstanceProfile `json:"instanceProfile,omitempty"`
}

type InstanceProfile struct {
	ARN  string `json:"arn,omitempty"`
	Name string `json:"name,omitempty"`
}
