package controllercontext

import "github.com/giantswarm/aws-operator/service/accountid"

type ContextService struct {
	AccountID *accountid.AccountID
}
