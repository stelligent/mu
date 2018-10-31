package common

// Roleset is a map of Role ARNs keyed by role type
type Roleset map[string]string

// RolesetUpserter for managing a roleset
type RolesetUpserter interface {
	UpsertCommonRoleset() error
	UpsertEnvironmentRoleset(environmentName string) error
	UpsertServiceRoleset(environmentName string, serviceName string, codeDeployBucket string, databaseName string) error
	UpsertPipelineRoleset(serviceName string, pipelineBucket string, codeDeployBucket string) error
}

// RolesetGetter for getting a roleset
type RolesetGetter interface {
	GetCommonRoleset() (Roleset, error)
	GetEnvironmentRoleset(environmentName string) (Roleset, error)
	GetEnvironmentProvider(environmentName string) (string, error)
	GetServiceRoleset(environmentName string, serviceName string) (Roleset, error)
	GetPipelineRoleset(serviceName string) (Roleset, error)
}

// RolesetDeleter for deleting a roleset
type RolesetDeleter interface {
	DeleteCommonRoleset() error
	DeleteEnvironmentRoleset(environmentName string) error
	DeleteServiceRoleset(environmentName string, serviceName string) error
	DeletePipelineRoleset(serviceName string) error
}

// RolesetManager composite of all roleset capabilities
type RolesetManager interface {
	RolesetGetter
	RolesetUpserter
	RolesetDeleter
}
