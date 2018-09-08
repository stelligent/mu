package common

import (
	"testing"

	"github.com/go-validator/validator"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfigServicePort(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configMax := Config{
		Service: Service{
			Port: 65536,
		},
	}
	config := Config{
		Service: Service{
			Port: 2,
		},
	}

	assert.NotNil(configMax.Validate())
	empty := configEmpty.Validate()
	assert.Nil(empty)
	assert.Nil(config.Validate())
}

func TestValidateConfigNamespace(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configDash := Config{
		Namespace: "-invalid",
	}
	configNumeric := Config{
		Namespace: "0invalid",
	}
	config := Config{
		Namespace: "c00l-stack-name",
	}

	assert.NotNil(configDash.Validate())
	assert.NotNil(configNumeric.Validate())
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}

func TestValidateConfigEnvironmentName(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configTooLong := Config{
		Environments: []Environment{
			Environment{
				Name: "jzdqtnkdprvavorzgaywhlbevajubhtcgciokrnehocaqltedhiaqmyostvwjdcm",
			},
		},
	}
	configInvalidChar := Config{
		Environments: []Environment{
			Environment{
				Name: "abc.123",
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				Name: "0test",
			},
		},
	}

	assert.NotNil(configTooLong.Validate())
	assert.NotNil(configInvalidChar.Validate())
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}

func TestValidateConfigEnvironmentLoadBalancer(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configTooLong := Config{
		Environments: []Environment{
			Environment{
				Loadbalancer: Loadbalancer{
					HostedZone: "eekphkulfncfvqkiumoxagowwpqawfohijlddjuyeicbsafkkfddoadxywhkudjnatjjocwrelftiefbexwidvohcoslcsxqvtdcwioxgzkavwtuiklbbakmfbroqcyszfmdyxgiiyhfsevzftsqbvjzwibzplmnwjusgslrjmfuyhhaqpirvrwgqvwuxkxegzblagxgvuuneecgeyjslcaamovanfedcodnlhbddibrtkjiaxkwjhcxlgqxbebm",
					Name:       "tfvrqoweaxnblkngzxsmhhcbapgqpzmwv",
				},
			},
		},
	}
	configInvalidChar := Config{
		Environments: []Environment{
			Environment{
				Loadbalancer: Loadbalancer{
					HostedZone: "abs=123",
					Name:       "abs.123",
				},
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				Loadbalancer: Loadbalancer{
					HostedZone: "0totally-cool.domain.com",
					Name:       "0totally-cooldomaincom",
				},
			},
		},
	}
	tooLong := configTooLong.Validate()
	tooLongErrs := tooLong.(validator.ErrorMap)

	invalidChar := configInvalidChar.Validate()
	invalidCharErrs := invalidChar.(validator.ErrorMap)

	assert.Equal(tooLongErrs["Environments[0].Loadbalancer.HostedZone"][0].Error(), "greater than max")
	assert.Equal(tooLongErrs["Environments[0].Loadbalancer.Name"][0].Error(), "greater than max")
	assert.Equal(invalidCharErrs["Environments[0].Loadbalancer.HostedZone"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].Loadbalancer.Name"][0].Error(), "regular expression mismatch")
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}

func TestValidateConfigEnvironmentCluster(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configInvalidChar := Config{
		Environments: []Environment{
			Environment{
				Cluster: Cluster{
					InstanceType:            "some-image",
					ImageID:                 "not-cool",
					TargetCPUReservation:    101,
					TargetMemoryReservation: 101,
					HTTPProxy:               "eekphkulfncfvqkiumoxagowwpqawfohijlddjuyeicbsafkkfddoadxywhkudjnatjjocwrelftiefbexwidvohcoslcsxqvtdcwioxgzkavwtuiklbbakmfbroqcyszfmdyxgiiyhfsevzftsqbvjzwibzplmnwjusgslrjmfuyhhaqpirvrwgqvwuxkxegzblagxgvuuneecgeyjslcaamovanfedcodnlhbddibrtkjiaxkwjhcxlgqxbebm",
				},
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				Cluster: Cluster{
					InstanceType:            "t2.small",
					ImageID:                 "ami-6cd6f714",
					TargetCPUReservation:    100,
					TargetMemoryReservation: 100,
					HTTPProxy:               "some.cool-domain.com/endpoint",
				},
			},
		},
	}

	invalidChar := configInvalidChar.Validate()
	invalidCharErrs := invalidChar.(validator.ErrorMap)

	assert.Equal(invalidCharErrs["Environments[0].Cluster.InstanceType"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].Cluster.ImageID"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].Cluster.TargetCPUReservation"][0].Error(), "greater than max")
	assert.Equal(invalidCharErrs["Environments[0].Cluster.TargetMemoryReservation"][0].Error(), "greater than max")
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}
func TestValidateConfigEnvironmentVpcTarget(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configInvalidChar := Config{
		Environments: []Environment{
			Environment{
				VpcTarget: VpcTarget{
					VpcID:             "some-vpc",
					InstanceSubnetIds: []string{"sub-1234safb"},
					ElbSubnetIds:      []string{"sub-1234safb"},
					Environment:       "n0t-valid_env",
					Namespace:         "n0t-valid_env",
				},
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				VpcTarget: VpcTarget{
					VpcID:             "vpc-i0m2a3g4e5",
					InstanceSubnetIds: []string{"subnet-s0u1b2n3e4t5"},
					ElbSubnetIds:      []string{"subnet-s0u1b2n3e4t5"},
					Environment:       "c00l-env",
					Namespace:         "c00l-env",
				},
			},
		},
	}

	invalidChar := configInvalidChar.Validate()
	invalidCharErrs := invalidChar.(validator.ErrorMap)

	assert.Equal(invalidCharErrs["Environments[0].VpcTarget.VpcID"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].VpcTarget.InstanceSubnetIds"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].VpcTarget.ElbSubnetIds"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].VpcTarget.Environment"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].VpcTarget.Namespace"][0].Error(), "regular expression mismatch")
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}
func TestValidateConfigEnvironmentRoles(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configInvalidChar := Config{
		Environments: []Environment{
			Environment{
				Roles: EnvironmentRoles{
					EcsInstance: "arn:aws:iam:000000000000:role/6",
				},
			},
		},
	}
	configTooLong := Config{
		Environments: []Environment{
			Environment{
				Roles: EnvironmentRoles{
					EcsInstance: "arn:aws:iam::000000000000:role/AbHQZQa4UUwfJm7hekFBf7sg92cUFy3SwxHRLV65uI7Ddp0zv5K4I5d3ZS9Y9U6kH",
				},
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				Roles: EnvironmentRoles{
					EcsInstance: "arn:aws:iam::000000000000:role/-+=,.@_4UUwfJm7hekFBf7sg92cUFy3wxHRLV65uI7Ddp0zv5K4I5d3ZS9Y9U6kH",
				},
			},
		},
	}

	invalidChar := configInvalidChar.Validate()
	invalidCharErrs := invalidChar.(validator.ErrorMap)

	tooLong := configTooLong.Validate()
	tooLongErrs := tooLong.(validator.ErrorMap)

	assert.Equal(invalidCharErrs["Environments[0].Roles.EcsInstance"][0].Error(), "regular expression mismatch")
	assert.Equal(tooLongErrs["Environments[0].Roles.EcsInstance"][0].Error(), "greater than max")
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}
