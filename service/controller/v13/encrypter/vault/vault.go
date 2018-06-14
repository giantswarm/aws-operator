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

	"github.com/giantswarm/aws-operator/service/controller/v13/encrypter"
	"github.com/giantswarm/aws-operator/service/controller/v13/key"
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
	}

	return e, nil
}

func (e *Encrypter) CreateKey(ctx context.Context, customObject v1alpha1.AWSConfig, _ string) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	keyName, err := e.EncryptionKey(ctx, customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	payload := &struct{}{}

	p := path.Join("transit", "keys", keyName)

	req, err := e.newPayloadRequest(p, payload)
	if err != nil {
		return microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return microerror.Mask(invalidHTTPStatusCodeError)
	}

	return nil
}

func (e *Encrypter) DeleteKey(ctx context.Context, keyName string) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	p := path.Join("transit", "keys", keyName)

	req, err := e.newRequest("DELETE", p)
	if err != nil {
		return microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return microerror.Mask(invalidHTTPStatusCodeError)
	}

	return nil
}

func (e *Encrypter) CurrentState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	state := encrypter.EncryptionKeyState{}

	err := e.ensureToken()
	if err != nil {
		return state, microerror.Mask(err)
	}

	keyName, err := e.EncryptionKey(ctx, customObject)
	if err != nil {
		return state, microerror.Mask(err)
	}

	p := path.Join("transit", "keys", keyName)

	req, err := e.newRequest("GET", p)
	if err != nil {
		return state, microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return state, microerror.Mask(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return state, microerror.Mask(invalidHTTPStatusCodeError)
	}

	state.KeyName = keyName

	return state, nil
}

func (e *Encrypter) DesiredState(ctx context.Context, customObject v1alpha1.AWSConfig) (encrypter.EncryptionKeyState, error) {
	state := encrypter.EncryptionKeyState{}

	keyName, err := e.EncryptionKey(ctx, customObject)
	if err != nil {
		return state, microerror.Mask(err)
	}

	state.KeyName = keyName

	return state, nil
}

func (e *Encrypter) EncryptionKey(ctx context.Context, customObject v1alpha1.AWSConfig) (string, error) {
	clusterID := key.ClusterID(customObject)

	return clusterID, nil
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
		return "", microerror.Mask(invalidHTTPStatusCodeError)
	}

	encryptResp := &EncryptResponse{}
	err = json.NewDecoder(resp.Body).Decode(encryptResp)
	if err != nil {
		return "", microerror.Mask(err)
	}

	ciphertext := encryptResp.Data.Ciphertext

	return ciphertext, nil
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
		return "", microerror.Mask(invalidHTTPStatusCodeError)
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

func (e *Encrypter) AddAWSIAMRoleToAuth(vaultRoleName, iamRoleARN string) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	p := path.Join("auth", "aws", "role", vaultRoleName)

	role, err := e.getAWSAuthRole(p)
	if err != nil {
		return microerror.Mask(err)
	}

	if role.BoundIAMRoleARN == "" {
		role.BoundIAMRoleARN = iamRoleARN
	} else if !strings.Contains(role.BoundIAMRoleARN, iamRoleARN) {
		role.BoundIAMRoleARN = fmt.Sprintf("%s,%s", role.BoundIAMRoleARN, iamRoleARN)
	}

	err = e.postAWSAuthRole(p, role)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (e *Encrypter) RemoveAWSIAMRoleFromAuth(vaultRoleName, iamRoleARN string) error {
	err := e.ensureToken()
	if err != nil {
		return microerror.Mask(err)
	}

	p := path.Join("auth", "aws", "role", vaultRoleName)

	role, err := e.getAWSAuthRole(p)
	if err != nil {
		return microerror.Mask(err)
	}

	iamRoles := strings.Split(role.BoundIAMRoleARN, ",")
	wantedIamRoles := []string{}
	for _, iamRole := range iamRoles {
		if iamRole != iamRoleARN {
			wantedIamRoles = append(wantedIamRoles, iamRole)
		}
	}
	role.BoundIAMRoleARN = strings.Join(wantedIamRoles, ",")

	err = e.postAWSAuthRole(p, role)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
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

	return resp.StatusCode == http.StatusOK
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
		return microerror.Mask(invalidHTTPStatusCodeError)
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

	req.Header.Set("Accept", "application/json")
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

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
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

	return strings.TrimSpace(string(responseData)), nil
}

func (e *Encrypter) getAWSAuthRole(path string) (*AWSAuthRole, error) {
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
		return nil, microerror.Mask(invalidHTTPStatusCodeError)
	}

	role := &AWSAuthRole{}
	err = json.NewDecoder(resp.Body).Decode(role)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return role, nil
}

func (e *Encrypter) postAWSAuthRole(path string, role *AWSAuthRole) error {
	req, err := e.newPayloadRequest(path, role)
	if err != nil {
		return microerror.Mask(err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return microerror.Mask(invalidHTTPStatusCodeError)
	}

	return nil
}
