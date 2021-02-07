package databricks

type ACLItem struct {
	Principal  string        `json:"principal"`
	Permission ACLPermission `json:"permission"`
}

type ACLPermission string

const (
	KindRead   ACLPermission = "READ"
	KindWrite  ACLPermission = "WRITE"
	KindManage ACLPermission = "MANAGE"
)

type ScopeBackendType string

const DatabricksBackend ScopeBackendType = "DATABRICKS"

type SecretScope struct {
	Name        string           `json:"name"`
	BackendType ScopeBackendType `json:"backend_type"`
}

type SecretMetadata struct {
	Key                  string `json:"key"`
	LastUpdatedTimestamp int64  `json:"last_updated_timestamp"`
}
