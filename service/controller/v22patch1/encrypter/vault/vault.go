package vault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/v22patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v22patch1/key"
)

const (
	decrypterVaultRole = "decrypter"
)

type Encrypter struct {
	httpClient *http.Client
	logger     micrologger.Logger

	address string
	base    *url.URL
	nonce   string
	token   string
}

type EncrypterConfig struct {
	Logger micrologger.Logger

	Address string
}

func NewEncrypter(c *EncrypterConfig) (*Encrypter, error) {
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.Address == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Address must not be empty", c)
	}

	base, err := url.Parse(c.Address + "/v1/")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// set client timeout to prevent leakages.
	httpClient := &http.Client{
		Timeout: time.Second * httpClientTimeout,
	}

	e := &Encrypter{
		httpClient: httpClient,
		logger:     c.Logger,

		address: c.Address,
		base:    base,
		// fixed nonce, so that we can reauthenticate from the same host.
		nonce: defaultNonce,
	}

	return e, nil
}

func (e *Encrypter) EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	err := e.ensureToken()
	if err != nil {
		return "", microerror.Mask(err)
	}

	keyName := e.keyName(customObject)

	p := path.Join("transit", "keys", keyName)

	req, err := e.newRequest("GET", p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", microerror.Mask(keyNotFoundError)
	} else if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", microerror.Maskf(invalidHTTPStatusCodeError, "want %d, got %d, response body: %q", http.StatusOK, resp.StatusCode, body)
	}

	return keyName, nil
}

func (e *Encrypter) Encrypt(ctx context.Context, key, plaintext string) (string, error) {
	err := e.ensureToken()
	if err != nil {
		return "", microerror.Mask(err)
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(plaintext))

	payload := &EncryptPayload{
		Plaintext: encoded,
	}

	p := path.Join("transit", "encrypt", key)

	req, err := e.newPayloadRequest(p, payload)
	if err != nil {
		return "", microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", microerror.Maskf(invalidHTTPStatusCodeError, "want 200, got %d, response body: %q", resp.StatusCode, body)
	}

	encryptResp := &EncryptResponse{}
	err = json.NewDecoder(resp.Body).Decode(encryptResp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	ciphertext := encryptResp.Data.Ciphertext

	return ciphertext, nil
}

func (e *Encrypter) EnsureCreatedAuthorizedIAMRoles(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var masterRoleARN string
	var workerRoleARN string
	{
		masterRoleARN = key.MasterRoleARN(customObject, cc.Status.TenantCluster.AWSAccountID)
		workerRoleARN = key.WorkerRoleARN(customObject, cc.Status.TenantCluster.AWSAccountID)
	}

	var roleData *AWSAuthRole
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding decrypter AWS auth role")

		roleData, err = e.getAuthAWSRole(decrypterVaultRole)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "found decrypter AWS auth role")
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "ensuring decrypter AWS auth role ARNs")

		currentARNs := roleData.BoundIAMRoleARN
		desiredARNs := stringSlice(roleData.BoundIAMRoleARN).Add(masterRoleARN, workerRoleARN)

		if len(currentARNs) != len(desiredARNs) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "updating decrypter AWS auth role")

			roleData.BoundIAMRoleARN = desiredARNs

			err = e.postAuthAWSRole(decrypterVaultRole, roleData)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			e.logger.LogCtx(ctx, "level", "debug", "message", "decrypter AWS auth role is up to date")
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "ensured decrypter AWS auth role ARNs")
	}

	return nil
}

func (e *Encrypter) EnsureCreatedEncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out encryption key")

		_, err := e.EncryptionKey(ctx, customObject)
		if IsKeyNotFound(err) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "did not find encryption key")

		} else if err != nil {
			return microerror.Mask(err)

		} else {
			e.logger.LogCtx(ctx, "level", "debug", "message", "found encryption key")
			e.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "creating encryption key")

		key := e.keyName(customObject)
		path := transitKeysPath(key)
		payload := &struct{}{}

		req, err := e.newPayloadRequest(path, payload)
		if err != nil {
			return microerror.Mask(err)
		}

		resp, err := e.httpClient.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			body, _ := ioutil.ReadAll(resp.Body)
			return microerror.Maskf(invalidHTTPStatusCodeError, "want 204, got %d, response body: %#q", resp.StatusCode, body)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "created encryption key")
	}

	return nil
}

func (e *Encrypter) EnsureDeletedAuthorizedIAMRoles(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	var masterRoleARN string
	var workerRoleARN string
	{
		masterRoleARN = key.MasterRoleARN(customObject, cc.Status.TenantCluster.AWSAccountID)
		workerRoleARN = key.WorkerRoleARN(customObject, cc.Status.TenantCluster.AWSAccountID)
	}

	var roleData *AWSAuthRole
	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out decrypter AWS auth role")

		roleData, err = e.getAuthAWSRole(decrypterVaultRole)
		if err != nil {
			return microerror.Mask(err)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "found decrypter AWS auth role")
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "ensuring deletion of decrypter AWS auth role ARNs")

		currentARNs := roleData.BoundIAMRoleARN
		desiredARNs := stringSlice(roleData.BoundIAMRoleARN).Delete(masterRoleARN, workerRoleARN)

		if len(currentARNs) != len(desiredARNs) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "updating decrypter AWS auth role")

			roleData.BoundIAMRoleARN = desiredARNs

			err = e.postAuthAWSRole(decrypterVaultRole, roleData)
			if err != nil {
				return microerror.Mask(err)
			}
		} else {
			e.logger.LogCtx(ctx, "level", "debug", "message", "decrypter AWS auth role is up to date")
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "ensured decrypter AWS auth role ARNs")
	}

	return nil
}

func (e *Encrypter) EnsureDeletedEncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "finding out encryption key")

		_, err := e.EncryptionKey(ctx, customObject)
		if IsKeyNotFound(err) {
			e.logger.LogCtx(ctx, "level", "debug", "message", "did not find encryption key")
			e.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil

		} else if err != nil {
			return microerror.Mask(err)

		} else {
			e.logger.LogCtx(ctx, "level", "debug", "message", "found encryption key")
		}
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "ensuring encryption key is deletable")

		key := e.keyName(customObject)
		path := transitKeysConfigPath(key)
		payload := &KeyConfigPayload{
			DeletionAllowed: true,
		}

		req, err := e.newPayloadRequest(path, payload)
		if err != nil {
			return microerror.Mask(err)
		}

		resp, err := e.httpClient.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			body, _ := ioutil.ReadAll(resp.Body)
			return microerror.Maskf(invalidHTTPStatusCodeError, "want 204, got %d, response body: %q", resp.StatusCode, body)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "ensured encryption key is deletable")
	}

	{
		e.logger.LogCtx(ctx, "level", "debug", "message", "deleting encryption key")

		key := e.keyName(customObject)
		path := transitKeysPath(key)

		req, err := e.newRequest("DELETE", path)
		if err != nil {
			return microerror.Mask(err)
		}

		resp, err := e.httpClient.Do(req)
		if err != nil {
			return microerror.Mask(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			body, _ := ioutil.ReadAll(resp.Body)
			return microerror.Maskf(invalidHTTPStatusCodeError, "want 204, got %d, response body: %q", resp.StatusCode, body)
		}

		e.logger.LogCtx(ctx, "level", "debug", "message", "deleted encryption key")
	}

	return nil
}

func (e *Encrypter) IsKeyNotFound(err error) bool {
	return IsKeyNotFound(err)
}

func (e *Encrypter) Decrypt(key, ciphertext string) (string, error) {
	err := e.ensureToken()
	if err != nil {
		return "", microerror.Mask(err)
	}

	payload := &DecryptPayload{
		Ciphertext: ciphertext,
	}

	p := path.Join("transit", "decrypt", key)

	req, err := e.newPayloadRequest(p, payload)
	if err != nil {
		return "", microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", microerror.Maskf(invalidHTTPStatusCodeError, "want 200, got %d, response body: %q", resp.StatusCode, body)
	}

	decryptResp := &DecryptResponse{}
	err = json.NewDecoder(resp.Body).Decode(decryptResp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	encoded := decryptResp.Data.Plaintext

	plaintext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(plaintext), nil
}

func (e *Encrypter) Address() string {
	return e.address
}

func (e *Encrypter) ensureToken() error {
	if e.isTokenValid() {
		return nil
	}

	err := e.login()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (e *Encrypter) isTokenValid() bool {
	p := path.Join("auth", "token", "lookup-self")

	req, err := e.newRequest("GET", p)
	if err != nil {
		e.logger.Log("level", "error", "message", fmt.Sprintf("could not create GET request for %q: %#v", p, err))
		return false
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		e.logger.Log("level", "error", "message", fmt.Sprintf("could not do %q request for %q: %#v", req.Method, req.URL, err))
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		e.logger.Log("level", "error", "message", fmt.Sprintf("invalid HTTP status code, want 200, got %d, response body: %q", resp.StatusCode, body))
		return false
	}
	return true
}

func (e *Encrypter) login() error {
	pkcs7, err := e.getPKCS7()
	if err != nil {
		return microerror.Mask(err)
	}

	payload := &LoginPayload{
		Role:  defaultRole,
		PKCS7: pkcs7,
		Nonce: e.nonce,
	}
	p := path.Join("auth", "aws", "login")

	req, err := e.newPayloadRequest(p, payload)
	if err != nil {
		return microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return microerror.Maskf(invalidHTTPStatusCodeError, "want 200, got %d, response body: %q", resp.StatusCode, body)
	}

	loginResp := &LoginResponse{}
	err = json.NewDecoder(resp.Body).Decode(loginResp)
	if err != nil {
		return microerror.Mask(err)
	}

	e.token = loginResp.Auth.ClientToken
	e.nonce = loginResp.Auth.Metadata.Nonce

	return nil
}

func (e *Encrypter) newRequest(method, path string) (*http.Request, error) {
	u := &url.URL{Path: path}
	dest := e.base.ResolveReference(u)

	var buf io.Reader

	req, err := http.NewRequest(method, dest.String(), buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if e.token != "" {
		req.Header.Set("X-Vault-Token", e.token)
	}

	return req, nil
}

func (e *Encrypter) newPayloadRequest(path string, payload interface{}) (*http.Request, error) {
	u := &url.URL{Path: path}
	dest := e.base.ResolveReference(u)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	buf := bytes.NewReader(b)

	req, err := http.NewRequest("POST", dest.String(), buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if e.token != "" {
		req.Header.Set("X-Vault-Token", e.token)
	}

	return req, nil
}

func (e *Encrypter) getPKCS7() (string, error) {
	response, err := http.Get(instanceIdentityPKCS7Endpoint)
	if err != nil {
		return "", microerror.Mask(err)
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return strings.Replace(string(responseData), "\n", "", -1), nil
}

func (e *Encrypter) getAuthAWSRole(name string) (*AWSAuthRole, error) {
	path := authAWSRolePath(name)

	req, err := e.newRequest("GET", path)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, microerror.Maskf(invalidHTTPStatusCodeError, "want 200, got %d, response body: %q", resp.StatusCode, body)
	}

	roleResponse := &AWSAuthRoleResponse{}
	err = json.NewDecoder(resp.Body).Decode(roleResponse)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &roleResponse.Data, nil
}

func (e *Encrypter) postAuthAWSRole(name string, data *AWSAuthRole) error {
	path := authAWSRolePath(name)

	req, err := e.newPayloadRequest(path, data)
	if err != nil {
		return microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return microerror.Maskf(invalidHTTPStatusCodeError, "want 200, got %d, response body: %q", resp.StatusCode, body)
	}

	return nil
}

func (e *Encrypter) keyName(customObject v1alpha1.AWSConfig) string {
	return key.ClusterID(customObject)
}

func authAWSRolePath(role string) string {
	return path.Join("auth", "aws", "role", role)
}

func transitKeysConfigPath(key string) string {
	return path.Join("transit", "keys", key, "config")
}

func transitKeysPath(key string) string {
	return path.Join("transit", "keys", key)
}
