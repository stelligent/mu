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
					HostedZone: "abs_123",
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
					InstanceType: "some-image",
					ImageID:      "not-cool",
				},
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				Cluster: Cluster{
					InstanceType: "t2.small",
					ImageID:      "ami-6cd6f714",
				},
			},
		},
	}

	invalidChar := configInvalidChar.Validate()
	invalidCharErrs := invalidChar.(validator.ErrorMap)

	assert.Equal(invalidCharErrs["Environments[0].Cluster.InstanceType"][0].Error(), "regular expression mismatch")
	assert.Equal(invalidCharErrs["Environments[0].Cluster.ImageID"][0].Error(), "regular expression mismatch")
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}
