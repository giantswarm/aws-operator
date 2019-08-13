package vault

const (
	httpClientTimeout = 5

	// instanceIdentityPKCS7Endpoint contains the fixed AWS endpoint to get an ec2
	// instance identity PKCS7 signature.
	instanceIdentityPKCS7Endpoint = "http://169.254.169.254/latest/dynamic/instance-identity/pkcs7"

	defaultRole = "encrypter"
	// fixed nonce, see https://www.vaultproject.io/api/auth/aws/index.html#nonce.
	defaultNonce = "aws-operator"
)

type LoginPayload struct {
	Role  string `json:"role"`
	PKCS7 string `json:"pkcs7"`
	Nonce string `json:"nonce,omitempty"`
}

type LoginResponse struct {
	Auth LoginAuthResponse `json:"auth"`
}

type LoginAuthResponse struct {
	Metadata    LoginAuthMetadataResponse `json:"metadata"`
	ClientToken string                    `json:"client_token"`
}

type LoginAuthMetadataResponse struct {
	Nonce string `json:"nonce"`
}

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

type EncryptPayload struct {
	Plaintext string `json:"plaintext"`
}

type EncryptResponse struct {
	Data EncryptResponseData `json:"data"`
}

type EncryptResponseData struct {
	Ciphertext string `json:"ciphertext"`
}

type DecryptPayload struct {
	Ciphertext string `json:"ciphertext"`
}

type DecryptResponse struct {
	Data DecryptResponseData `json:"data"`
}

type DecryptResponseData struct {
	Plaintext string `json:"plaintext"`
}

type AWSAuthRoleResponse struct {
	Data AWSAuthRole `json:"data"`
}

type AWSAuthRole struct {
	BoundIAMRoleARN []string `json:"bound_iam_role_arn"`
}

type KeyConfigPayload struct {
	DeletionAllowed bool `json:"deletion_allowed"`
}
