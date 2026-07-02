export interface ClientRole {
  id: string; // Keycloak role ID
  name: string; // full name: clientId:resource:action
  description?: string;
  composite?: boolean;
}
