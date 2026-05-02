package domain

import "strings"

type FederatedIdentity struct {
	Username, Domain string
}

func ParseFederatedIdentity(s string) (FederatedIdentity, error) {
	s = strings.TrimSpace(s)
	username, domain, err := parseAtParts(s, "federated identity", "username")
	if err != nil {
		return FederatedIdentity{}, err
	}
	return FederatedIdentity{Username: username, Domain: domain}, nil
}

func (f FederatedIdentity) String() string {
	return f.Username + "@" + f.Domain
}

func (f FederatedIdentity) IsLocal(localDomain string) bool {
	return f.Domain == localDomain
}
