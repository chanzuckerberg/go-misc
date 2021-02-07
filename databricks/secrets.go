package databricks

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

func (o *DBClientOption) CreateSecretScope(ctx context.Context, scope, initialManagePrincipal string) error {
	data := map[string]string{
		"scope":                    scope,
		"initial_manage_principal": initialManagePrincipal,
	}

	httpRequest, err := o.craftRequest(ctx, "/secrets/scopes/create", "POST", data)
	if err != nil {
		return errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to execute http request")
	}

	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	return nil
}

func (o *DBClientOption) DeleteSecretScope(ctx context.Context, scope string) error {
	data := map[string]string{
		"scope": scope,
	}
	httpRequest, err := o.craftRequest(ctx, "/secrets/scopes/delete", "POST", data)
	if err != nil {
		return errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to execute http request")
	}

	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	return nil
}

func (o *DBClientOption) ListSecretScopes(ctx context.Context) ([]SecretScope, error) {
	data := map[string]string{}

	var scopes []SecretScope

	httpRequest, err := o.craftRequest(ctx, "/secrets/scopes/list", "GET", data)
	if err != nil {
		return scopes, errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return scopes, errors.Wrap(err, "Unable to execute http request")
	}
	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return scopes, errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return scopes, fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	// Unmarshal to interface
	var respMap map[string]interface{}

	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return scopes, errors.Wrapf(err, "Unable to unmarshal http response's body to map[string]interface. Got %T", resp.Body)
	}

	scopesInterface, ok := respMap["scopes"]
	if !ok {
		return scopes, errors.Wrap(err, "ListSecretScopes() response body doesn't contain 'scopes' key")
	}

	scopes, ok = scopesInterface.([]SecretScope)
	if !ok {
		return scopes, errors.Wrap(err, "Unable to convert scopesInterface value to type []SecretScope")
	}

	return scopes, errors.Wrap(err, "Unable to unmarshal http request body to SecretScopes array")
}

func (o *DBClientOption) PutSecret(ctx context.Context, scope, key string) error {
	// Get the secret as a bytes so we can plug in bytes_value into request
	data := map[string]string{
		"scope": scope,
		"key":   key,
	}
	httpRequest, err := o.craftRequest(ctx, "/secrets/put", "POST", data)
	if err != nil {
		return errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to execute http request")
	}

	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	return nil
}

func (o *DBClientOption) DeleteSecret(ctx context.Context, scope, key string) error {
	data := map[string]string{
		"scope": scope,
		"key":   key,
	}
	httpRequest, err := o.craftRequest(ctx, "/secrets/delete", "POST", data)
	if err != nil {
		return errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to execute http request")
	}

	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to read http response body")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	// User must have write or manage permission on the secret scope
	return nil
}

func (o *DBClientOption) ListSecrets(ctx context.Context, scope string) ([]SecretMetadata, error) {
	data := map[string]string{
		"scope": scope,
	}

	var secrets []SecretMetadata

	httpRequest, err := o.craftRequest(ctx, "/secrets/list", "GET", data)
	if err != nil {
		return secrets, errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return secrets, errors.Wrap(err, "Unable to execute http request")
	}
	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return secrets, errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return secrets, fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	var respMap map[string]interface{}
	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return secrets, errors.Wrapf(err, "Unable to unmarshal http response's body to map[string]interface. Got %T", resp.Body)
	}

	secretsInterface, ok := respMap["secrets"]
	if !ok {
		return secrets, errors.Wrap(err, "ListSecrets() response body doesn't contain 'secrets' key")
	}

	secrets, ok = secretsInterface.([]SecretMetadata)
	if !ok {
		return secrets, errors.Wrap(err, "Unable to convert secretsInterface value to type []SecretMetadata")
	}

	return secrets, errors.Wrap(err, "Unable to unmarshal http request body to Secrets array")
}

func (o *DBClientOption) PutSecretACL(ctx context.Context, scope, principal string, permission ACLPermission) error {
	// Must have kindManage permission to invoke this API
	data := map[string]string{
		"scope":      scope,
		"principal":  principal,
		"permission": string(permission),
	}
	httpRequest, err := o.craftRequest(ctx, "/secrets/acls/put", "POST", data)
	if err != nil {
		return errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to execute http request")
	}

	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	return nil
}

func (o *DBClientOption) DeleteSecretACL(ctx context.Context, scope, principal string) error {
	data := map[string]string{
		"scope":     scope,
		"principal": principal,
	}
	httpRequest, err := o.craftRequest(ctx, "/secrets/acls/delete", "POST", data)
	if err != nil {
		return errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrap(err, "Unable to execute http request")
	}

	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	return nil
}

func (o *DBClientOption) GetSecretACL(ctx context.Context, scope, principal string) (ACLItem, error) {
	data := map[string]string{
		"scope":     scope,
		"principal": principal,
	}
	var acl ACLItem
	httpRequest, err := o.craftRequest(ctx, "/secrets/acls/get", "GET", data)
	if err != nil {
		return acl, errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return acl, errors.Wrap(err, "Unable to execute http request")
	}
	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return acl, errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return acl, fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()

	err = dec.Decode(&acl)

	return acl, errors.Wrap(err, "Unable to unmarshal http request body to AclItem type")
}

func (o *DBClientOption) ListSecretACLs(ctx context.Context, scope string) ([]ACLItem, error) {
	data := map[string]string{
		"scope": scope,
	}

	var acls []ACLItem

	httpRequest, err := o.craftRequest(ctx, "/secrets/acls/list", "GET", data)
	if err != nil {
		return acls, errors.Wrap(err, "Unable to create http request")
	}

	httpClient := o.getHTTPClient()

	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return acls, errors.Wrap(err, "Unable to execute http request")
	}
	defer resp.Body.Close()
	// We should be able to read the request body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return acls, errors.Wrap(err, "Unable to read http response body")
	}

	if resp.StatusCode != 200 {
		return acls, fmt.Errorf("Response from server (%d) %s", resp.StatusCode, string(body))
	}

	var respMap map[string]interface{}
	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return acls, errors.Wrapf(err, "Unable to unmarshal http response's body to map[string]interface. Got %T", resp.Body)
	}

	aclsInterface, ok := respMap["items"]
	if !ok {
		return acls, errors.Wrap(err, "ListSecretACLs() response body doesn't contain 'acls' key")
	}

	acls, ok = aclsInterface.([]ACLItem)
	if !ok {
		return acls, errors.Wrap(err, "Unable to convert aclsInterface value to type []AclItem")
	}

	return acls, errors.Wrap(err, "Unable to unmarshal http request body to Secrets array")
}
