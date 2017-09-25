package common

// Roleset is a map of Role ARNs keyed by role type
type Roleset map[string]string

// RolesetUpserter for managing a roleset
type RolesetUpserter interface {
	UpsertCommonRoleset() (Roleset, error)
	UpsertEnvironmentRoleset(environmentName string) (Roleset, error)
	UpsertServiceRoleset(environmentName string, serviceName string) (Roleset, error)
	UpsertPipelineRoleset(serviceName string) (Roleset, error)
}

// RolesetGetter for getting a roleset
type RolesetGetter interface {
	GetCommonRoleset() (Roleset, error)
	GetUpsertEnvironmentRoleset(environmentName string) (Roleset, error)
	GetServiceRoleset(environmentName string, serviceName string) (Roleset, error)
	GetPipelineRoleset(serviceName string) (Roleset, error)
}

// RolesetManager composite of all roleset capabilities
type RolesetManager interface {
	RolesetGetter
	RolesetUpserter
}
