package sessions

import "github.com/moh-sso-dashboard/internal/keycloak"

func toSessionResponse(session keycloak.Session) SessionResponse {
	return SessionResponse{
		ID:         session.ID,
		IPAddress:  session.IP,
		Start:      session.Start,
		LastAccess: session.LastAccess,
		Clients:    session.Clients,
	}
}

func toSessionResponses(sessions []keycloak.Session) []SessionResponse {
	out := make([]SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		out = append(out, toSessionResponse(session))
	}
	return out
}
