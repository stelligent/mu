package common

// CatalogUpserter for upserting catalogs
type CatalogUpserter interface {
	SetProductVersions(productID string, productVersions map[string]string) error
}

// CatalogProvisioner for provisioning products
type CatalogProvisioner interface {
	UpsertProvisionedProduct(productID string, version string, name string, params map[string]string) error
}

// CatalogTerminator for terminating catalogs
type CatalogTerminator interface {
	TerminateProvisionedProducts(productID string) error
}

// CatalogManager composite of all catalog capabilities
type CatalogManager interface {
	CatalogUpserter
	CatalogProvisioner
	CatalogTerminator
}
