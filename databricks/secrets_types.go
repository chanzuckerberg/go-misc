package databricks

type AclItem struct {
	Principal  string        `json:"principal"`
	Permission AclPermission `json:"permission"`
}

type AclPermission string

const (
	KindRead   AclPermission = "READ"
	KindWrite  AclPermission = "WRITE"
	KindManage AclPermission = "MANAGE"
)

type ScopeBackendType string

const DatabricksBackend ScopeBackendType = "DATABRICKS"

type SecretScope struct {
	Name         string           `json:"name"`
	Backend_type ScopeBackendType `json:"backend_type"`
}

type SecretMetadata struct {
	Key                    string `json:"key"`
	Last_updated_timestamp int64  `json:"last_updated_timestamp"`
}
