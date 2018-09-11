package common

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-validator/validator"
	"github.com/stretchr/testify/assert"
)

func TestValidateResourceID(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(validateResourceID("ami-51537029", "ami"))
	assert.Nil(validateResourceID([]string{
		"ami-6cd6f714",
		"ami-0d0d4f8ea9e3965a6",
	}, "ami"))
	assert.Nil(validateResourceID([]string{
		"subnet-003cbb4b",
		"subnet-0135cf12e41d1a961",
	}, "subnet"))
	assert.Nil(validateResourceID([]string{
		"vpc-b0c134c8",
		"vpc-0d6d40bc880d2fc5d",
	}, "vpc"))
	assert.NotNil(validateResourceID([]string{
		"vpc-b0c134c8",
		"vpc-0d6d40bc880d2fc5d",
	}, "subnet"))
}

func TestValidateCIDR(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(validateCIDR("0.0.0.0/0", ""))
	assert.Nil(validateCIDR("192.168.1.1/32", ""))
	assert.NotNil(validateCIDR("0.0.0.0", ""))
	assert.NotNil(validateCIDR("0.0.0.0/100", ""))
}

func TestValidateRoleARN(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(validateRoleARN("arn:aws:iam::000000000000:role/aws-service-role/ecs.amazonaws.com/AWSServiceRoleForECS", ""))
	assert.Nil(validateRoleARN("arn:aws:iam::000000000000:role/aggauth-ConfigRecorderRole-15L", ""))
	assert.NotNil(validateRoleARN("arn:aws:iam::000000000000:role", ""))
}
func TestValidateInstanceType(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(validateInstanceType("t2.nano", ""))
	assert.Nil(validateInstanceType("t3.2xlarge", ""))
	assert.Nil(validateInstanceType("m5d.xlarge", ""))
	assert.Nil(validateInstanceType("m5d.24xlarge", ""))
	assert.Nil(validateInstanceType("m5d.24xlarge", ""))
	assert.Nil(validateInstanceType("x1e.32xlarge", ""))
	assert.Nil(validateInstanceType("i3.metal", ""))
	assert.Nil(validateInstanceType("p3.16xlarge", ""))
	assert.Nil(validateInstanceType("db.m4.large", ""))
	assert.Nil(validateInstanceType("db.m3.2xlarge", ""))
	assert.Nil(validateInstanceType("db.x1e.16xlarge", ""))
	assert.Nil(validateInstanceType("db.t2.small", ""))
	assert.NotNil(validateInstanceType("a.bar", ""))
	assert.NotNil(validateInstanceType("a2.foo", ""))
}
func TestValidateURL(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(validateURL("foo.BAR/foo-bar_", ""))
	assert.Nil(validateURL("/", ""))
	assert.Nil(validateURL("/ping", ""))
	assert.Nil(validateURL("a", ""))
	assert.NotNil(validateURL("foo@bar.foo-bar", ""))
	assert.NotNil(validateURL("http://foo.bar", ""))
}
func TestValidateLeadingAlphaNumericDash(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(validateLeadingAlphaNumericDash("FOO-bar", ""))
	assert.Nil(validateLeadingAlphaNumericDash("00-foo-bar", ""))
	assert.NotNil(validateLeadingAlphaNumericDash("f", ""))
	assert.NotNil(validateLeadingAlphaNumericDash("-foobar", ""))
	assert.NotNil(validateLeadingAlphaNumericDash("foo.bar", ""))
}
func TestValidateAlphaNumericDash(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(validateAlphaNumericDash("FOO-bar", ""))
	assert.NotNil(validateAlphaNumericDash("00-foo-bar", ""))
	assert.NotNil(validateAlphaNumericDash("f", ""))
	assert.NotNil(validateAlphaNumericDash("-foobar", ""))
	assert.NotNil(validateAlphaNumericDash("foo.bar", ""))
}
func TestRegexpLength(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(regexpLength("foo", "", 99))
	assert.Nil(regexpLength("foo", "foo", 99))
	assert.Nil(regexpLength("foo", "foo?", 99))

	tooLongErr := regexpLength("foo", "", 1)
	assert.Equal(tooLongErr.Error(), "greater than max")
}

func TestSomeString(t *testing.T) {
	assert := assert.New(t)

	returnV := func(v interface{}, param string) error {
		st := reflect.ValueOf(v)
		value := st.String()
		return errors.New(value)
	}

	returnNil := func(v interface{}, param string) error {
		return nil
	}

	matchString := someString(reflect.ValueOf([]string{"foo"}), "foo", returnV)
	assert.Equal(matchString.Error(), "foo")
	assert.Nil(someString(reflect.ValueOf([]string{"foo"}), "bar", returnNil))
	assert.NotNil(someString(reflect.ValueOf([]string{"foo"}), "bar", returnV))
}
func TestRegex(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(regex("foo", "foo"))
	assert.NotNil(regex("foo", "bar"))
}

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
					Instance: "arn:aws:iam:000000000000:role/6",
				},
			},
		},
	}
	configTooLong := Config{
		Environments: []Environment{
			Environment{
				Roles: EnvironmentRoles{
					Instance: "arn:aws:iam::000000000000:role/AbHQZQa4UUwfJm7hekFBf7sg92cUFy3SwxHRLV65uI7Ddp0zv5K4I5d3ZS9Y9U6kH",
				},
			},
		},
	}
	config := Config{
		Environments: []Environment{
			Environment{
				Roles: EnvironmentRoles{
					Instance: "arn:aws:iam::000000000000:role/-+=,.@_4UUwfJm7hekFBf7sg92cUFy3wxHRLV65uI7Ddp0zv5K4I5d3ZS9Y9U6kH",
				},
			},
		},
	}

	invalidChar := configInvalidChar.Validate()
	invalidCharErrs := invalidChar.(validator.ErrorMap)

	tooLong := configTooLong.Validate()
	tooLongErrs := tooLong.(validator.ErrorMap)

	assert.Equal(invalidCharErrs["Environments[0].Roles.Instance"][0].Error(), "regular expression mismatch")
	assert.Equal(tooLongErrs["Environments[0].Roles.Instance"][0].Error(), "greater than max")
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}
