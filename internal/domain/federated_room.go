package domain

import "strings"

type FederatedRoom struct {
	RoomName, Domain string
}

func ParseFederatedRoom(s string) (FederatedRoom, error) {
	s = strings.TrimSpace(s)
	roomName, domain, err := parseAtParts(s, "federated room", "room name")
	if err != nil {
		return FederatedRoom{}, err
	}
	return FederatedRoom{RoomName: roomName, Domain: domain}, nil
}

func (f FederatedRoom) String() string {
	return f.RoomName + "@" + f.Domain
}
