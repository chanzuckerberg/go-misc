package rotate

import (
	"os"
	"testing"

	"github.com/chanzuckerberg/aws-oidc/pkg/util"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xinsnake/databricks-sdk-golang/aws/models"
)

type testSecretAPI struct {
	secretScopes []models.SecretScope
	secrets      map[models.SecretScope]testSecret
	acls         []*models.AclItem
}

type testSecret map[string][]byte

func newTestSecretAPI() *testSecretAPI {
	return &testSecretAPI{
		secretScopes: []models.SecretScope{},
		secrets:      map[models.SecretScope]testSecret{},
		acls:         []*models.AclItem{},
	}
}

func (t *testSecretAPI) ListSecretScopes() ([]models.SecretScope, error) {
	return t.secretScopes, nil
}

func (t *testSecretAPI) PutSecretACL(scope string, principal string, permission models.AclPermission) error {
	t.acls = append(t.acls, &models.AclItem{
		Principal:  principal,
		Permission: &permission,
	})

	return nil
}

func (t *testSecretAPI) CreateSecretScope(scope string, initialManagePrincipal string) error {
	t.secretScopes = append(t.secretScopes, models.SecretScope{Name: scope})
	return nil
}

func (t *testSecretAPI) PutSecret(bytesValue []byte, scopeName string, key string) error {
	currentSecretScopes, err := t.ListSecretScopes()
	if err != nil {
		return err
	}

	for _, scopeItem := range currentSecretScopes {
		if scopeItem.Name == scopeName {
			t.secrets[scopeItem] = testSecret{key: bytesValue}
			return nil
		}
	}

	return errors.New("Scope doesn't exist")
}

func TestUpdateDatabricksNewScopes(t *testing.T) {
	r := require.New(t)
	defer util.ResetEnv(os.Environ())
	acctName := "testAccount"
	testSnowflakeCredentials := []*snowflakeUserCredentials{
		{
			user:            "user1",
			role:            "role1",
			pem_private_key: "privkey1",
		},
		{
			user:            "user2",
			role:            "role2",
			pem_private_key: "privkey2",
		},
	}
	dummySecretClient := newTestSecretAPI()
	uniqueScopes := []string{"newscope1", "newscope2"}

	dummyScopeList, err := dummySecretClient.ListSecretScopes()
	r.NoError(err)
	r.Len(dummyScopeList, 0)
	for i, testCredential := range testSnowflakeCredentials {
		err := updateDatabricks(uniqueScopes[i], acctName, testCredential, dummySecretClient)
		r.NoError(err)
	}

	dummyScopeList, err = dummySecretClient.ListSecretScopes()
	r.NoError(err)
	r.Len(dummyScopeList, 2)

	// Writing the credentials again with new users, same scopes
	// the scope list should be the same length
	// secrets map should be expanded ()
	testSnowflakeCredentials[0].user = "user3"
	testSnowflakeCredentials[0].pem_private_key = "privkey3"
	testSnowflakeCredentials[1].user = "user4"
	testSnowflakeCredentials[1].pem_private_key = "privkey4"
	for i, testCredential := range testSnowflakeCredentials {
		err := updateDatabricks(uniqueScopes[i], acctName, testCredential, dummySecretClient)
		r.NoError(err)
	}
	dummyScopeList, err = dummySecretClient.ListSecretScopes()
	r.NoError(err)
	r.Len(dummyScopeList, 2)
}
