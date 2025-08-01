package gocloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// GetQueryParams converts the struct to map[string]string
// The fields tags must have `json:"<name>,string,omitempty"` format for all types, except strings
// The string fields must have: `json:"<name>,omitempty"`. The `json:"<name>,string,omitempty"` tag for string field
// will add additional double quotes.
// "string" tag allows to convert the non-string fields of a structure to map[string]string.
// "omitempty" allows to skip the fields with default values.
func GetQueryParams(s interface{}) (map[string]string, error) {
	// if obj, ok := s.(GetGroupsParams); ok {
	// 	obj.OnMarshal()
	// 	s = obj
	// }
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var res map[string]string
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// StringOrArray represents a value that can either be a string or an array of strings
type StringOrArray []string

// UnmarshalJSON unmarshals a string or an array object from a JSON array or a JSON string
func (s *StringOrArray) UnmarshalJSON(data []byte) error {
	if len(data) > 1 && data[0] == '[' {
		var obj []string
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		*s = StringOrArray(obj)
		return nil
	}

	var obj string
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	*s = StringOrArray([]string{obj})
	return nil
}

// MarshalJSON converts the array of strings to a JSON array or JSON string if there is only one item in the array
func (s *StringOrArray) MarshalJSON() ([]byte, error) {
	if len(*s) == 1 {
		return json.Marshal([]string(*s)[0])
	}
	return json.Marshal([]string(*s))
}

// EnforcedString can be used when the expected value is string but Keycloak in some cases gives you mixed types
type EnforcedString string

// UnmarshalJSON modify data as string before json unmarshal
func (s *EnforcedString) UnmarshalJSON(data []byte) error {
	if data[0] != '"' {
		// Escape unescaped quotes
		data = bytes.ReplaceAll(data, []byte(`"`), []byte(`\"`))
		data = bytes.ReplaceAll(data, []byte(`\\"`), []byte(`\"`))

		// Wrap data in quotes
		data = append([]byte(`"`), data...)
		data = append(data, []byte(`"`)...)
	}

	var val string
	err := json.Unmarshal(data, &val)
	*s = EnforcedString(val)
	return err
}

// MarshalJSON return json marshal
func (s *EnforcedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(*s)
}

// APIErrType is a field containing more specific API error types
// that may be checked by the receiver.
type APIErrType string

const (
	// APIErrTypeUnknown is for API errors that are not strongly
	// typed.
	APIErrTypeUnknown APIErrType = "unknown"

	// APIErrTypeInvalidGrant corresponds with Keycloak's
	// OAuthErrorException due to "invalid_grant".
	APIErrTypeInvalidGrant = "oauth: invalid grant"
)

// ParseAPIErrType is a convenience method for returning strongly
// typed API errors.
func ParseAPIErrType(err error) APIErrType {
	if err == nil {
		return APIErrTypeUnknown
	}
	switch {
	case strings.Contains(err.Error(), "invalid_grant"):
		return APIErrTypeInvalidGrant
	default:
		return APIErrTypeUnknown
	}
}

// APIError holds message and statusCode for api errors
type APIError struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Type    APIErrType `json:"type"`
}

// Error stringifies the APIError
func (apiError APIError) Error() string {
	return apiError.Message
}

// CertResponseKey is returned by the certs endpoint.
// JSON Web Key structure is described here:
// https://self-issued.info/docs/draft-ietf-jose-json-web-key.html#JWKContents
type CertResponseKey struct {
	Kid     *string   `json:"kid,omitempty"`
	Kty     *string   `json:"kty,omitempty"`
	Alg     *string   `json:"alg,omitempty"`
	Use     *string   `json:"use,omitempty"`
	N       *string   `json:"n,omitempty"`
	E       *string   `json:"e,omitempty"`
	X       *string   `json:"x,omitempty"`
	Y       *string   `json:"y,omitempty"`
	Crv     *string   `json:"crv,omitempty"`
	KeyOps  *[]string `json:"key_ops,omitempty"`
	X5u     *string   `json:"x5u,omitempty"`
	X5c     *[]string `json:"x5c,omitempty"`
	X5t     *string   `json:"x5t,omitempty"`
	X5tS256 *string   `json:"x5t#S256,omitempty"`
}

// CertResponse is returned by the certs endpoint
type CertResponse struct {
	Keys *[]CertResponseKey `json:"keys,omitempty"`
}

// IssuerResponse is returned by the issuer endpoint
type IssuerResponse struct {
	Realm           *string `json:"realm,omitempty"`
	PublicKey       *string `json:"public_key,omitempty"`
	TokenService    *string `json:"token-service,omitempty"`
	AccountService  *string `json:"account-service,omitempty"`
	TokensNotBefore *int    `json:"tokens-not-before,omitempty"`
}

// ResourcePermission represents a permission granted to a resource
type ResourcePermission struct {
	RSID           *string   `json:"rsid,omitempty"`
	ResourceID     *string   `json:"resource_id,omitempty"`
	RSName         *string   `json:"rsname,omitempty"`
	Scopes         *[]string `json:"scopes,omitempty"`
	ResourceScopes *[]string `json:"resource_scopes,omitempty"`
}

// PermissionResource represents a resources asscoiated with a permission
type PermissionResource struct {
	ResourceID   *string `json:"_id,omitempty"`
	ResourceName *string `json:"name,omitempty"`
}

// PermissionScope represents scopes associated with a permission
type PermissionScope struct {
	ScopeID   *string `json:"id,omitempty"`
	ScopeName *string `json:"name,omitempty"`
}

// IntroSpectTokenResult is returned when a token was checked
type IntroSpectTokenResult struct {
	Permissions *[]ResourcePermission `json:"permissions,omitempty"`
	Exp         *int                  `json:"exp,omitempty"`
	Nbf         *int                  `json:"nbf,omitempty"`
	Iat         *int                  `json:"iat,omitempty"`
	Aud         *StringOrArray        `json:"aud,omitempty"`
	Active      *bool                 `json:"active,omitempty"`
	AuthTime    *int                  `json:"auth_time,omitempty"`
	Jti         *string               `json:"jti,omitempty"`
	Type        *string               `json:"typ,omitempty"`
	Azp         *string               `json:"azp,omitempty"`
}

// User represents the Keycloak User Structure
type User struct {
	ID                         *string                     `json:"id,omitempty"`
	CreatedTimestamp           *int64                      `json:"createdTimestamp,omitempty"`
	Username                   *string                     `json:"username,omitempty"`
	Enabled                    *bool                       `json:"enabled,omitempty"`
	Totp                       *bool                       `json:"totp,omitempty"`
	EmailVerified              *bool                       `json:"emailVerified,omitempty"`
	FirstName                  *string                     `json:"firstName,omitempty"`
	LastName                   *string                     `json:"lastName,omitempty"`
	Email                      *string                     `json:"email,omitempty"`
	FederationLink             *string                     `json:"federationLink,omitempty"`
	Attributes                 *map[string][]string        `json:"attributes,omitempty"`
	DisableableCredentialTypes *[]interface{}              `json:"disableableCredentialTypes,omitempty"`
	RequiredActions            *[]string                   `json:"requiredActions,omitempty"`
	Access                     *map[string]bool            `json:"access,omitempty"`
	ClientRoles                *map[string][]string        `json:"clientRoles,omitempty"`
	RealmRoles                 *[]string                   `json:"realmRoles,omitempty"`
	Groups                     *[]string                   `json:"groups,omitempty"`
	ServiceAccountClientID     *string                     `json:"serviceAccountClientId,omitempty"`
	Credentials                *[]CredentialRepresentation `json:"credentials,omitempty"`
	NotBefore                  *int64                      `json:"notBefore,omitempty"`
	Origin                     *string                     `json:"origin,omitempty"`
	Self                       *string                     `json:"self,omitempty"`
}

// SetPasswordRequest sets a new password
type SetPasswordRequest struct {
	Type      *string `json:"type,omitempty"`
	Temporary *bool   `json:"temporary,omitempty"`
	Password  *string `json:"value,omitempty"`
}

// Component is a component
type Component struct {
	ID              *string              `json:"id,omitempty"`
	Name            *string              `json:"name,omitempty"`
	ProviderID      *string              `json:"providerId,omitempty"`
	ProviderType    *string              `json:"providerType,omitempty"`
	ParentID        *string              `json:"parentId,omitempty"`
	ComponentConfig *map[string][]string `json:"config,omitempty"`
	SubType         *string              `json:"subType,omitempty"`
}

// KeyStoreConfig holds the keyStoreConfig
type KeyStoreConfig struct {
	ActiveKeys *ActiveKeys `json:"active,omitempty"`
	Key        *[]Key      `json:"keys,omitempty"`
}

// ActiveKeys holds the active keys
type ActiveKeys struct {
	HS256 *string `json:"HS256,omitempty"`
	RS256 *string `json:"RS256,omitempty"`
	AES   *string `json:"AES,omitempty"`
}

// Key is a key
type Key struct {
	ProviderID       *string `json:"providerId,omitempty"`
	ProviderPriority *int    `json:"providerPriority,omitempty"`
	Kid              *string `json:"kid,omitempty"`
	Status           *string `json:"status,omitempty"`
	Type             *string `json:"type,omitempty"`
	Algorithm        *string `json:"algorithm,omitempty"`
	PublicKey        *string `json:"publicKey,omitempty"`
	Certificate      *string `json:"certificate,omitempty"`
}

// Attributes holds Attributes
type Attributes struct {
	LDAPENTRYDN *[]string `json:"LDAP_ENTRY_DN,omitempty"`
	LDAPID      *[]string `json:"LDAP_ID,omitempty"`
}

// Access represents access
type Access struct {
	ManageGroupMembership *bool `json:"manageGroupMembership,omitempty"`
	View                  *bool `json:"view,omitempty"`
	MapRoles              *bool `json:"mapRoles,omitempty"`
	Impersonate           *bool `json:"impersonate,omitempty"`
	Manage                *bool `json:"manage,omitempty"`
}

// UserGroup is a UserGroup
type UserGroup struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	Path *string `json:"path,omitempty"`
}

// GetUsersParams represents the optional parameters for getting users
type GetUsersParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Email               *string `json:"email,omitempty"`
	EmailVerified       *bool   `json:"emailVerified,string,omitempty"`
	Enabled             *bool   `json:"enabled,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	FirstName           *string `json:"firstName,omitempty"`
	IDPAlias            *string `json:"idpAlias,omitempty"`
	IDPUserID           *string `json:"idpUserId,omitempty"`
	LastName            *string `json:"lastName,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
	Username            *string `json:"username,omitempty"`
}

// GetComponentsParams represents the optional parameters for getting components
type GetComponentsParams struct {
	Name         *string `json:"name,omitempty"`
	ProviderType *string `json:"provider,omitempty"`
	ParentID     *string `json:"parent,omitempty"`
}

// ExecuteActionsEmail represents parameters for executing action emails
type ExecuteActionsEmail struct {
	UserID      *string   `json:"-"`
	ClientID    *string   `json:"client_id,omitempty"`
	Lifespan    *int      `json:"lifespan,string,omitempty"`
	RedirectURI *string   `json:"redirect_uri,omitempty"`
	Actions     *[]string `json:"-"`
}

// SendVerificationMailParams is being used to send verification params
type SendVerificationMailParams struct {
	ClientID    *string
	RedirectURI *string
}

// Group is a Group
type Group struct {
	ID            *string              `json:"id,omitempty"`
	Name          *string              `json:"name,omitempty"`
	Path          *string              `json:"path,omitempty"`
	SubGroups     *[]Group             `json:"subGroups,omitempty"`
	SubGroupCount *int                 `json:"subGroupCount,omitempty"`
	Attributes    *map[string][]string `json:"attributes,omitempty"`
	Access        *map[string]bool     `json:"access,omitempty"`
	ClientRoles   *map[string][]string `json:"clientRoles,omitempty"`
	RealmRoles    *[]string            `json:"realmRoles,omitempty"`
	ParentID      *string              `json:"parentId,omitempty"`
}

// GroupsCount represents the groups count response from keycloak
type GroupsCount struct {
	Count int `json:"count,omitempty"`
}

// GetGroupsParams represents the optional parameters for getting groups
type GetGroupsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Full                *bool   `json:"full,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
}

// GetChildGroupsParams represents the optional parameters for getting child groups
type GetChildGroupsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Search              *string `json:"search,omitempty"`
}

// MarshalJSON is a custom json marshaling function to automatically set the Full and BriefRepresentation properties
// for backward compatibility
func (obj GetGroupsParams) MarshalJSON() ([]byte, error) {
	type Alias GetGroupsParams
	a := (Alias)(obj)
	if a.BriefRepresentation != nil {
		a.Full = BoolP(!*a.BriefRepresentation)
	} else if a.Full != nil {
		a.BriefRepresentation = BoolP(!*a.Full)
	}
	return json.Marshal(a)
}

// CompositesRepresentation represents the composite roles of a role
type CompositesRepresentation struct {
	Client *map[string][]string `json:"client,omitempty"`
	Realm  *[]string            `json:"realm,omitempty"`
}

// Role is a role
type Role struct {
	ID                 *string                   `json:"id,omitempty"`
	Name               *string                   `json:"name,omitempty"`
	ScopeParamRequired *bool                     `json:"scopeParamRequired,omitempty"`
	Composite          *bool                     `json:"composite,omitempty"`
	Composites         *CompositesRepresentation `json:"composites,omitempty"`
	ClientRole         *bool                     `json:"clientRole,omitempty"`
	ContainerID        *string                   `json:"containerId,omitempty"`
	Description        *string                   `json:"description,omitempty"`
	Attributes         *map[string][]string      `json:"attributes,omitempty"`
	Access             *map[string]bool          `json:"access,omitempty"`
}

// GetRoleParams represents the optional parameters for getting roles
type GetRoleParams struct {
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Search              *string `json:"search,omitempty"`
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
}

// ClientMappingsRepresentation is a client role mappings
type ClientMappingsRepresentation struct {
	ID       *string `json:"id,omitempty"`
	Client   *string `json:"client,omitempty"`
	Mappings *[]Role `json:"mappings,omitempty"`
}

// MappingsRepresentation is a representation of role mappings
type MappingsRepresentation struct {
	ClientMappings map[string]*ClientMappingsRepresentation `json:"clientMappings,omitempty"`
	RealmMappings  *[]Role                                  `json:"realmMappings,omitempty"`
}

// ClientScope is a ClientScope
type ClientScope struct {
	ID                    *string                `json:"id,omitempty"`
	Name                  *string                `json:"name,omitempty"`
	Type                  *string                `json:"type,omitempty"`
	Description           *string                `json:"description,omitempty"`
	Protocol              *string                `json:"protocol,omitempty"`
	ClientScopeAttributes *ClientScopeAttributes `json:"attributes,omitempty"`
	ProtocolMappers       *[]ProtocolMappers     `json:"protocolMappers,omitempty"`
}

// ClientScopeAttributes are attributes of client scopes
type ClientScopeAttributes struct {
	ConsentScreenText      *string `json:"consent.screen.text,omitempty"`
	DisplayOnConsentScreen *string `json:"display.on.consent.screen,omitempty"`
	IncludeInTokenScope    *string `json:"include.in.token.scope,omitempty"`
}

// ProtocolMappers are protocolmappers
type ProtocolMappers struct {
	ID                    *string                `json:"id,omitempty"`
	Name                  *string                `json:"name,omitempty"`
	Protocol              *string                `json:"protocol,omitempty"`
	ProtocolMapper        *string                `json:"protocolMapper,omitempty"`
	ConsentRequired       *bool                  `json:"consentRequired,omitempty"`
	ProtocolMappersConfig *ProtocolMappersConfig `json:"config,omitempty"`
}

// ProtocolMappersConfig is a config of a protocol mapper
type ProtocolMappersConfig struct {
	UserinfoTokenClaim                 *string `json:"userinfo.token.claim,omitempty"`
	UserAttribute                      *string `json:"user.attribute,omitempty"`
	IDTokenClaim                       *string `json:"id.token.claim,omitempty"`
	AccessTokenClaim                   *string `json:"access.token.claim,omitempty"`
	ClaimName                          *string `json:"claim.name,omitempty"`
	ClaimValue                         *string `json:"claim.value,omitempty"`
	JSONTypeLabel                      *string `json:"jsonType.label,omitempty"`
	Multivalued                        *string `json:"multivalued,omitempty"`
	AggregateAttrs                     *string `json:"aggregate.attrs,omitempty"`
	UsermodelClientRoleMappingClientID *string `json:"usermodel.clientRoleMapping.clientId,omitempty"`
	IncludedClientAudience             *string `json:"included.client.audience,omitempty"`
	FullPath                           *string `json:"full.path,omitempty"`
	AttributeName                      *string `json:"attribute.name,omitempty"`
	AttributeNameFormat                *string `json:"attribute.nameformat,omitempty"`
	Single                             *string `json:"single,omitempty"`
	Script                             *string `json:"script,omitempty"`
	AddOrganizationAttributes          *string `json:"addOrganizationAttributes,omitempty"`
	AddOrganizationID                  *string `json:"addOrganizationId,omitempty"`
}

// Client is a ClientRepresentation
type Client struct {
	Access                               *map[string]interface{}         `json:"access,omitempty"`
	AdminURL                             *string                         `json:"adminUrl,omitempty"`
	Attributes                           *map[string]string              `json:"attributes,omitempty"`
	AuthenticationFlowBindingOverrides   *map[string]string              `json:"authenticationFlowBindingOverrides,omitempty"`
	AuthorizationServicesEnabled         *bool                           `json:"authorizationServicesEnabled,omitempty"`
	AuthorizationSettings                *ResourceServerRepresentation   `json:"authorizationSettings,omitempty"`
	BaseURL                              *string                         `json:"baseUrl,omitempty"`
	BearerOnly                           *bool                           `json:"bearerOnly,omitempty"`
	ClientAuthenticatorType              *string                         `json:"clientAuthenticatorType,omitempty"`
	ClientID                             *string                         `json:"clientId,omitempty"`
	ClientTemplate                       *string                         `json:"clientTemplate,omitempty"`
	ConsentRequired                      *bool                           `json:"consentRequired,omitempty"`
	DefaultClientScopes                  *[]string                       `json:"defaultClientScopes,omitempty"`
	DefaultRoles                         *[]string                       `json:"defaultRoles,omitempty"`
	Description                          *string                         `json:"description,omitempty"`
	DirectAccessGrantsEnabled            *bool                           `json:"directAccessGrantsEnabled,omitempty"`
	Enabled                              *bool                           `json:"enabled,omitempty"`
	FrontChannelLogout                   *bool                           `json:"frontchannelLogout,omitempty"`
	FullScopeAllowed                     *bool                           `json:"fullScopeAllowed,omitempty"`
	ID                                   *string                         `json:"id,omitempty"`
	ImplicitFlowEnabled                  *bool                           `json:"implicitFlowEnabled,omitempty"`
	Name                                 *string                         `json:"name,omitempty"`
	NodeReRegistrationTimeout            *int32                          `json:"nodeReRegistrationTimeout,omitempty"`
	NotBefore                            *int32                          `json:"notBefore,omitempty"`
	OptionalClientScopes                 *[]string                       `json:"optionalClientScopes,omitempty"`
	Origin                               *string                         `json:"origin,omitempty"`
	Protocol                             *string                         `json:"protocol,omitempty"`
	ProtocolMappers                      *[]ProtocolMapperRepresentation `json:"protocolMappers,omitempty"`
	PublicClient                         *bool                           `json:"publicClient,omitempty"`
	RedirectURIs                         *[]string                       `json:"redirectUris,omitempty"`
	RegisteredNodes                      *map[string]int                 `json:"registeredNodes,omitempty"`
	RegistrationAccessToken              *string                         `json:"registrationAccessToken,omitempty"`
	RootURL                              *string                         `json:"rootUrl,omitempty"`
	Secret                               *string                         `json:"secret,omitempty"`
	ServiceAccountsEnabled               *bool                           `json:"serviceAccountsEnabled,omitempty"`
	StandardFlowEnabled                  *bool                           `json:"standardFlowEnabled,omitempty"`
	SurrogateAuthRequired                *bool                           `json:"surrogateAuthRequired,omitempty"`
	UseTemplateConfig                    *bool                           `json:"useTemplateConfig,omitempty"`
	UseTemplateMappers                   *bool                           `json:"useTemplateMappers,omitempty"`
	UseTemplateScope                     *bool                           `json:"useTemplateScope,omitempty"`
	WebOrigins                           *[]string                       `json:"webOrigins,omitempty"`
	AlwaysDisplayInConsole               *bool                           `json:"alwaysDisplayInConsole,omitempty"`
	BackchannelLogoutRevokeOfflineTokens *bool                           `json:"backchannelLogoutRevokeOfflineTokens,omitempty"`
	BackchannelLogoutSessionRequired     *bool                           `json:"backchannelLogoutSessionRequired,omitempty"`
	BackchannelLogoutURL                 *string                         `json:"backchannelLogoutUrl,omitempty"`
	ClientSessionIdleTimeout             *int                            `json:"clientSessionIdleTimeout,omitempty"`
	ClientSessionMaxLifespan             *int                            `json:"clientSessionMaxLifespan,omitempty"`
	ClientOfflineSessionIdleTimeout      *int                            `json:"clientOfflineSessionIdleTimeout,omitempty"`
	ClientOfflineSessionMaxLifespan      *int                            `json:"clientOfflineSessionMaxLifespan,omitempty"`
	Logo                                 *string                         `json:"logo,omitempty"`
	PolicyURI                            *string                         `json:"policyUri,omitempty"`
	SAMLASSERTIONSIGNATURE               *bool                           `json:"saml.assertion.signature,omitempty"`
	SAMLAutodetect                       *bool                           `json:"saml.autodetect,omitempty"`
	SAMLClientSignature                  *bool                           `json:"saml.client.signature,omitempty"`
	SAMLEncrypt                          *bool                           `json:"saml.encrypt,omitempty"`
	SAMLForcePostBinding                 *bool                           `json:"saml.force.post.binding,omitempty"`
	SAMLMultiValuedRoles                 *bool                           `json:"saml.multivalued.roles,omitempty"`
	SAMLServerSignature                  *bool                           `json:"saml.server.signature,omitempty"`
	SAMLSignatureAlgorithm               *string                         `json:"saml.signature.algorithm,omitempty"`
	TosURI                               *string                         `json:"tosUri,omitempty"`
}

// ResourceServerRepresentation represents the resources of a Server
type ResourceServerRepresentation struct {
	AllowRemoteResourceManagement *bool                     `json:"allowRemoteResourceManagement,omitempty"`
	ClientID                      *string                   `json:"clientId,omitempty"`
	ID                            *string                   `json:"id,omitempty"`
	Name                          *string                   `json:"name,omitempty"`
	Policies                      *[]PolicyRepresentation   `json:"policies,omitempty"`
	PolicyEnforcementMode         *PolicyEnforcementMode    `json:"policyEnforcementMode,omitempty"`
	Resources                     *[]ResourceRepresentation `json:"resources,omitempty"`
	Scopes                        *[]ScopeRepresentation    `json:"scopes,omitempty"`
	DecisionStrategy              *DecisionStrategy         `json:"decisionStrategy,omitempty"`
}

// RoleDefinition represents a role in a RolePolicyRepresentation
type RoleDefinition struct {
	ID       *string `json:"id,omitempty"`
	Private  *bool   `json:"private,omitempty"`
	Required *bool   `json:"required,omitempty"`
}

// AdapterConfiguration represents adapter configuration of a client
type AdapterConfiguration struct {
	Realm            *string     `json:"realm"`
	AuthServerURL    *string     `json:"auth-server-url"`
	SSLRequired      *string     `json:"ssl-required"`
	Resource         *string     `json:"resource"`
	Credentials      interface{} `json:"credentials"`
	ConfidentialPort *int        `json:"confidential-port"`
}

// PolicyEnforcementMode is an enum type for PolicyEnforcementMode of ResourceServerRepresentation
type PolicyEnforcementMode string

// PolicyEnforcementMode values
var (
	ENFORCING  = PolicyEnforcementModeP("ENFORCING")
	PERMISSIVE = PolicyEnforcementModeP("PERMISSIVE")
	DISABLED   = PolicyEnforcementModeP("DISABLED")
)

// Logic is an enum type for policy logic
type Logic string

// Logic values
var (
	POSITIVE = LogicP("POSITIVE")
	NEGATIVE = LogicP("NEGATIVE")
)

// DecisionStrategy is an enum type for DecisionStrategy of PolicyRepresentation
type DecisionStrategy string

// DecisionStrategy values
var (
	AFFIRMATIVE = DecisionStrategyP("AFFIRMATIVE")
	UNANIMOUS   = DecisionStrategyP("UNANIMOUS")
	CONSENSUS   = DecisionStrategyP("CONSENSUS")
)

// AbstractPolicyRepresentation is the base representation for all policies
type AbstractPolicyRepresentation struct {
	Config           *map[string]string `json:"config,omitempty"`
	DecisionStrategy *DecisionStrategy  `json:"decisionStrategy,omitempty"`
	Description      *string            `json:"description,omitempty"`
	ID               *string            `json:"id,omitempty"`
	Logic            *Logic             `json:"logic,omitempty"`
	Name             *string            `json:"name,omitempty"`
	Owner            *string            `json:"owner,omitempty"`
	Policies         *[]string          `json:"policies,omitempty"`
	Resources        *[]string          `json:"resources,omitempty"`
	Scopes           *[]string          `json:"scopes,omitempty"`
	Type             *string            `json:"type,omitempty"`
}

// PolicyRepresentation is a representation of a Policy
type PolicyRepresentation struct {
	Config           *map[string]string `json:"config,omitempty"`
	DecisionStrategy *DecisionStrategy  `json:"decisionStrategy,omitempty"`
	Description      *string            `json:"description,omitempty"`
	ID               *string            `json:"id,omitempty"`
	Logic            *Logic             `json:"logic,omitempty"`
	Name             *string            `json:"name,omitempty"`
	Owner            *string            `json:"owner,omitempty"`
	Policies         *[]string          `json:"policies,omitempty"`
	Resources        *[]string          `json:"resources,omitempty"`
	Scopes           *[]string          `json:"scopes,omitempty"`
	Type             *string            `json:"type,omitempty"`
	RolePolicyRepresentation
	JSPolicyRepresentation
	ClientPolicyRepresentation
	TimePolicyRepresentation
	UserPolicyRepresentation
	AggregatedPolicyRepresentation
	GroupPolicyRepresentation
}

// ToConfig converts embedded policy-specific fields to Config format for Keycloak 26+ compatibility
func (p *PolicyRepresentation) ToConfig() {
	if p.Config == nil {
		p.Config = &map[string]string{}
	}

	config := *p.Config

	// Convert ClientPolicyRepresentation to Config
	if p.ClientPolicyRepresentation.Clients != nil && len(*p.ClientPolicyRepresentation.Clients) > 0 {
		clients := make([]string, len(*p.ClientPolicyRepresentation.Clients))
		for i, client := range *p.ClientPolicyRepresentation.Clients {
			clients[i] = fmt.Sprintf(`"%s"`, client)
		}
		config["clients"] = fmt.Sprintf("[%s]", strings.Join(clients, ","))
	}

	// Convert RolePolicyRepresentation to Config
	if p.RolePolicyRepresentation.Roles != nil && len(*p.RolePolicyRepresentation.Roles) > 0 {
		roles := make([]string, len(*p.RolePolicyRepresentation.Roles))
		for i, role := range *p.RolePolicyRepresentation.Roles {
			if role.ID != nil {
				roles[i] = fmt.Sprintf(`{"id":"%s","required":%t}`, *role.ID, role.Required != nil && *role.Required)
			}
		}
		config["roles"] = fmt.Sprintf("[%s]", strings.Join(roles, ","))
	}

	// Convert JSPolicyRepresentation to Config
	if p.JSPolicyRepresentation.Code != nil {
		config["code"] = *p.JSPolicyRepresentation.Code
	}

	// Convert UserPolicyRepresentation to Config
	if p.UserPolicyRepresentation.Users != nil && len(*p.UserPolicyRepresentation.Users) > 0 {
		users := make([]string, len(*p.UserPolicyRepresentation.Users))
		for i, user := range *p.UserPolicyRepresentation.Users {
			users[i] = fmt.Sprintf(`"%s"`, user)
		}
		config["users"] = fmt.Sprintf("[%s]", strings.Join(users, ","))
	}

	// Convert AggregatedPolicyRepresentation to Config
	if p.AggregatedPolicyRepresentation.Policies != nil && len(*p.AggregatedPolicyRepresentation.Policies) > 0 {
		policies := make([]string, len(*p.AggregatedPolicyRepresentation.Policies))
		for i, policy := range *p.AggregatedPolicyRepresentation.Policies {
			policies[i] = fmt.Sprintf(`"%s"`, policy)
		}
		config["applyPolicies"] = fmt.Sprintf("[%s]", strings.Join(policies, ","))
	}

	// Convert GroupPolicyRepresentation to Config
	if p.GroupPolicyRepresentation.Groups != nil && len(*p.GroupPolicyRepresentation.Groups) > 0 {
		groups := make([]string, len(*p.GroupPolicyRepresentation.Groups))
		for i, group := range *p.GroupPolicyRepresentation.Groups {
			if group.ID != nil {
				required := "false"
				if group.ExtendChildren != nil && *group.ExtendChildren {
					required = "true"
				}
				groups[i] = fmt.Sprintf(`{"id":"%s","extendChildren":%s}`, *group.ID, required)
			}
		}
		config["groups"] = fmt.Sprintf("[%s]", strings.Join(groups, ","))
	}
	if p.GroupPolicyRepresentation.GroupsClaim != nil {
		config["groupsClaim"] = *p.GroupPolicyRepresentation.GroupsClaim
	}

	// Convert TimePolicyRepresentation to Config
	if p.TimePolicyRepresentation.NotBefore != nil {
		config["nbf"] = *p.TimePolicyRepresentation.NotBefore
	}
	if p.TimePolicyRepresentation.NotOnOrAfter != nil {
		config["noa"] = *p.TimePolicyRepresentation.NotOnOrAfter
	}
	if p.TimePolicyRepresentation.DayMonth != nil {
		config["dayMonth"] = *p.TimePolicyRepresentation.DayMonth
	}
	if p.TimePolicyRepresentation.DayMonthEnd != nil {
		config["dayMonthEnd"] = *p.TimePolicyRepresentation.DayMonthEnd
	}
	if p.TimePolicyRepresentation.Month != nil {
		config["month"] = *p.TimePolicyRepresentation.Month
	}
	if p.TimePolicyRepresentation.MonthEnd != nil {
		config["monthEnd"] = *p.TimePolicyRepresentation.MonthEnd
	}
	if p.TimePolicyRepresentation.Year != nil {
		config["year"] = *p.TimePolicyRepresentation.Year
	}
	if p.TimePolicyRepresentation.YearEnd != nil {
		config["yearEnd"] = *p.TimePolicyRepresentation.YearEnd
	}
	if p.TimePolicyRepresentation.Hour != nil {
		config["hour"] = *p.TimePolicyRepresentation.Hour
	}
	if p.TimePolicyRepresentation.HourEnd != nil {
		config["hourEnd"] = *p.TimePolicyRepresentation.HourEnd
	}
	if p.TimePolicyRepresentation.Minute != nil {
		config["minute"] = *p.TimePolicyRepresentation.Minute
	}
	if p.TimePolicyRepresentation.MinuteEnd != nil {
		config["minuteEnd"] = *p.TimePolicyRepresentation.MinuteEnd
	}

	p.Config = &config
}

// RolePolicyRepresentation represents role based policies
type RolePolicyRepresentation struct {
	Roles *[]RoleDefinition `json:"roles,omitempty"`
}

// JSPolicyRepresentation represents js based policies
type JSPolicyRepresentation struct {
	Code *string `json:"code,omitempty"`
}

// ClientPolicyRepresentation represents client based policies
type ClientPolicyRepresentation struct {
	Clients *[]string `json:"clients,omitempty"`
}

// TimePolicyRepresentation represents time based policies
type TimePolicyRepresentation struct {
	NotBefore    *string `json:"notBefore,omitempty"`
	NotOnOrAfter *string `json:"notOnOrAfter,omitempty"`
	DayMonth     *string `json:"dayMonth,omitempty"`
	DayMonthEnd  *string `json:"dayMonthEnd,omitempty"`
	Month        *string `json:"month,omitempty"`
	MonthEnd     *string `json:"monthEnd,omitempty"`
	Year         *string `json:"year,omitempty"`
	YearEnd      *string `json:"yearEnd,omitempty"`
	Hour         *string `json:"hour,omitempty"`
	HourEnd      *string `json:"hourEnd,omitempty"`
	Minute       *string `json:"minute,omitempty"`
	MinuteEnd    *string `json:"minuteEnd,omitempty"`
}

// UserPolicyRepresentation represents user based policies
type UserPolicyRepresentation struct {
	Users *[]string `json:"users,omitempty"`
}

// AggregatedPolicyRepresentation represents aggregated policies
type AggregatedPolicyRepresentation struct {
	Policies *[]string `json:"policies,omitempty"`
}

// GroupPolicyRepresentation represents group based policies
type GroupPolicyRepresentation struct {
	Groups      *[]GroupDefinition `json:"groups,omitempty"`
	GroupsClaim *string            `json:"groupsClaim,omitempty"`
}

// GroupDefinition represents a group in a GroupPolicyRepresentation
type GroupDefinition struct {
	ID             *string `json:"id,omitempty"`
	Path           *string `json:"path,omitempty"`
	ExtendChildren *bool   `json:"extendChildren,omitempty"`
}

// ResourceRepresentation is a representation of a Resource
type ResourceRepresentation struct {
	ID                 *string                      `json:"_id,omitempty"` // TODO: is marked "_optional" in template, input error or deliberate?
	Attributes         *map[string][]string         `json:"attributes,omitempty"`
	DisplayName        *string                      `json:"displayName,omitempty"`
	IconURI            *string                      `json:"icon_uri,omitempty"` // TODO: With "_" because that's how it's written down in the template
	Name               *string                      `json:"name,omitempty"`
	Owner              *ResourceOwnerRepresentation `json:"owner,omitempty"`
	OwnerManagedAccess *bool                        `json:"ownerManagedAccess,omitempty"`
	ResourceScopes     *[]ScopeRepresentation       `json:"resource_scopes,omitempty"`
	Scopes             *[]ScopeRepresentation       `json:"scopes,omitempty"`
	Type               *string                      `json:"type,omitempty"`
	URIs               *[]string                    `json:"uris,omitempty"`
}

// ResourceOwnerRepresentation represents a resource's owner
type ResourceOwnerRepresentation struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// ScopeRepresentation is a represents a Scope
type ScopeRepresentation struct {
	DisplayName *string                   `json:"displayName,omitempty"`
	IconURI     *string                   `json:"iconUri,omitempty"`
	ID          *string                   `json:"id,omitempty"`
	Name        *string                   `json:"name,omitempty"`
	Policies    *[]PolicyRepresentation   `json:"policies,omitempty"`
	Resources   *[]ResourceRepresentation `json:"resources,omitempty"`
}

// ProtocolMapperRepresentation represents....
type ProtocolMapperRepresentation struct {
	Config          *map[string]string `json:"config,omitempty"`
	ID              *string            `json:"id,omitempty"`
	Name            *string            `json:"name,omitempty"`
	Protocol        *string            `json:"protocol,omitempty"`
	ProtocolMapper  *string            `json:"protocolMapper,omitempty"`
	ConsentRequired *bool              `json:"consentRequired,omitempty"`
}

// GetClientsParams represents the query parameters
type GetClientsParams struct {
	ClientID             *string `json:"clientId,omitempty"`
	ViewableOnly         *bool   `json:"viewableOnly,string,omitempty"`
	First                *int    `json:"first,string,omitempty"`
	Max                  *int    `json:"max,string,omitempty"`
	Search               *bool   `json:"search,string,omitempty"`
	SearchableAttributes *string `json:"q,omitempty"`
}

// UserInfoAddress is representation of the address sub-filed of UserInfo
// https://openid.net/specs/openid-connect-core-1_0.html#AddressClaim
type UserInfoAddress struct {
	Formatted     *string `json:"formatted,omitempty"`
	StreetAddress *string `json:"street_address,omitempty"`
	Locality      *string `json:"locality,omitempty"`
	Region        *string `json:"region,omitempty"`
	PostalCode    *string `json:"postal_code,omitempty"`
	Country       *string `json:"country,omitempty"`
}

// UserInfo is returned by the userinfo endpoint
// https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
type UserInfo struct {
	Sub                 *string          `json:"sub,omitempty"`
	Name                *string          `json:"name,omitempty"`
	GivenName           *string          `json:"given_name,omitempty"`
	FamilyName          *string          `json:"family_name,omitempty"`
	MiddleName          *string          `json:"middle_name,omitempty"`
	Nickname            *string          `json:"nickname,omitempty"`
	PreferredUsername   *string          `json:"preferred_username,omitempty"`
	Profile             *string          `json:"profile,omitempty"`
	Picture             *string          `json:"picture,omitempty"`
	Website             *string          `json:"website,omitempty"`
	Email               *string          `json:"email,omitempty"`
	EmailVerified       *bool            `json:"email_verified,omitempty"`
	Gender              *string          `json:"gender,omitempty"`
	ZoneInfo            *string          `json:"zoneinfo,omitempty"`
	Locale              *string          `json:"locale,omitempty"`
	PhoneNumber         *string          `json:"phone_number,omitempty"`
	PhoneNumberVerified *bool            `json:"phone_number_verified,omitempty"`
	Address             *UserInfoAddress `json:"address,omitempty"`
	UpdatedAt           *int             `json:"updated_at,omitempty"`
}

// RolesRepresentation represents the roles of a realm
type RolesRepresentation struct {
	Client *map[string][]Role `json:"client,omitempty"`
	Realm  *[]Role            `json:"realm,omitempty"`
}

// RealmRepresentation represents a realm
type RealmRepresentation struct {
	AccessCodeLifespan                                        *int                              `json:"accessCodeLifespan,omitempty"`
	AccessCodeLifespanLogin                                   *int                              `json:"accessCodeLifespanLogin,omitempty"`
	AccessCodeLifespanUserAction                              *int                              `json:"accessCodeLifespanUserAction,omitempty"`
	AccessTokenLifespan                                       *int                              `json:"accessTokenLifespan,omitempty"`
	AccessTokenLifespanForImplicitFlow                        *int                              `json:"accessTokenLifespanForImplicitFlow,omitempty"`
	AccountTheme                                              *string                           `json:"accountTheme,omitempty"`
	ActionTokenGeneratedByAdminLifespan                       *int                              `json:"actionTokenGeneratedByAdminLifespan,omitempty"`
	ActionTokenGeneratedByUserLifespan                        *int                              `json:"actionTokenGeneratedByUserLifespan,omitempty"`
	AdminEventsDetailsEnabled                                 *bool                             `json:"adminEventsDetailsEnabled,omitempty"`
	AdminEventsEnabled                                        *bool                             `json:"adminEventsEnabled,omitempty"`
	AdminTheme                                                *string                           `json:"adminTheme,omitempty"`
	Attributes                                                *map[string]string                `json:"attributes,omitempty"`
	AuthenticationFlows                                       *[]interface{}                    `json:"authenticationFlows,omitempty"`
	AuthenticatorConfig                                       *[]interface{}                    `json:"authenticatorConfig,omitempty"`
	BruteForceProtected                                       *bool                             `json:"bruteForceProtected,omitempty"`
	BruteForceStrategy                                        *string                           `json:"bruteForceStrategy,omitempty"`
	BrowserFlow                                               *string                           `json:"browserFlow,omitempty"`
	BrowserSecurityHeaders                                    *map[string]string                `json:"browserSecurityHeaders,omitempty"`
	ClientOfflineSessionIdleTimeout                           *int                              `json:"clientOfflineSessionIdleTimeout,omitempty"`
	ClientOfflineSessionMaxLifespan                           *int                              `json:"clientOfflineSessionMaxLifespan,omitempty"`
	ClientAuthenticationFlow                                  *string                           `json:"clientAuthenticationFlow,omitempty"`
	ClientPolicies                                            *map[string][]interface{}         `json:"clientPolicies,omitempty"`
	ClientProfiles                                            *map[string][]interface{}         `json:"clientProfiles,omitempty"`
	ClientScopeMappings                                       *map[string][]interface{}         `json:"clientScopeMappings,omitempty"`
	ClientScopes                                              *[]ClientScope                    `json:"clientScopes,omitempty"`
	ClientSessionIdleTimeout                                  *int                              `json:"clientSessionIdleTimeout,omitempty"`
	ClientSessionMaxLifespan                                  *int                              `json:"clientSessionMaxLifespan,omitempty"`
	Clients                                                   *[]Client                         `json:"clients,omitempty"`
	Components                                                *map[string][]Component           `json:"components,omitempty"`
	DefaultDefaultClientScopes                                *[]string                         `json:"defaultDefaultClientScopes,omitempty"`
	DefaultGroups                                             *[]string                         `json:"defaultGroups,omitempty"`
	DefaultLocale                                             *string                           `json:"defaultLocale,omitempty"`
	DefaultOptionalClientScopes                               *[]string                         `json:"defaultOptionalClientScopes,omitempty"`
	DefaultRole                                               *Role                             `json:"defaultRole,omitempty"`
	DefaultRoles                                              *[]string                         `json:"defaultRoles,omitempty"`
	DefaultSignatureAlgorithm                                 *string                           `json:"defaultSignatureAlgorithm,omitempty"`
	DirectGrantFlow                                           *string                           `json:"directGrantFlow,omitempty"`
	DisplayName                                               *string                           `json:"displayName,omitempty"`
	DisplayNameHTML                                           *string                           `json:"displayNameHtml,omitempty"`
	DuplicateEmailsAllowed                                    *bool                             `json:"duplicateEmailsAllowed,omitempty"`
	EditUsernameAllowed                                       *bool                             `json:"editUsernameAllowed,omitempty"`
	EmailTheme                                                *string                           `json:"emailTheme,omitempty"`
	Enabled                                                   *bool                             `json:"enabled,omitempty"`
	EnabledEventTypes                                         *[]string                         `json:"enabledEventTypes,omitempty"`
	EventsEnabled                                             *bool                             `json:"eventsEnabled,omitempty"`
	EventsListeners                                           *[]string                         `json:"eventsListeners,omitempty"`
	FailureFactor                                             *int                              `json:"failureFactor,omitempty"`
	FederatedUsers                                            *[]interface{}                    `json:"federatedUsers,omitempty"`
	Groups                                                    *[]Group                          `json:"groups,omitempty"`
	ID                                                        *string                           `json:"id,omitempty"`
	IdentityProviderMappers                                   *[]IdentityProviderMapper         `json:"identityProviderMappers,omitempty"`
	IdentityProviders                                         *[]IdentityProviderRepresentation `json:"identityProviders,omitempty"`
	InternationalizationEnabled                               *bool                             `json:"internationalizationEnabled,omitempty"`
	KeycloakVersion                                           *string                           `json:"keycloakVersion,omitempty"`
	LoginTheme                                                *string                           `json:"loginTheme,omitempty"`
	LocalizationTexts                                         *map[string]map[string]string     `json:"localizationTexts,omitempty"`
	LoginWithEmailAllowed                                     *bool                             `json:"loginWithEmailAllowed,omitempty"`
	MaxDeltaTimeSeconds                                       *int                              `json:"maxDeltaTimeSeconds,omitempty"`
	MaxFailureWaitSeconds                                     *int                              `json:"maxFailureWaitSeconds,omitempty"`
	MaxTemporaryLockouts                                      *int                              `json:"maxTemporaryLockouts,omitempty"`
	MinimumQuickLoginWaitSeconds                              *int                              `json:"minimumQuickLoginWaitSeconds,omitempty"`
	NotBefore                                                 *int                              `json:"notBefore,omitempty"`
	OAuth2DeviceCodeLifespan                                  *int                              `json:"oauth2DeviceCodeLifespan,omitempty"`
	OAuth2DevicePollingInterval                               *int                              `json:"oauth2DevicePollingInterval,omitempty"`
	OfflineSessionIdleTimeout                                 *int                              `json:"offlineSessionIdleTimeout,omitempty"`
	OfflineSessionMaxLifespan                                 *int                              `json:"offlineSessionMaxLifespan,omitempty"`
	OfflineSessionMaxLifespanEnabled                          *bool                             `json:"offlineSessionMaxLifespanEnabled,omitempty"`
	OrganizationsEnabled                                      *bool                             `json:"organizationsEnabled,omitempty"`
	OTPPolicyAlgorithm                                        *string                           `json:"otpPolicyAlgorithm,omitempty"`
	OTPPolicyCodeReusable                                     *bool                             `json:"otpPolicyCodeReusable,omitempty"`
	OTPPolicyDigits                                           *int                              `json:"otpPolicyDigits,omitempty"`
	OTPPolicyInitialCounter                                   *int                              `json:"otpPolicyInitialCounter,omitempty"`
	OTPPolicyLookAheadWindow                                  *int                              `json:"otpPolicyLookAheadWindow,omitempty"`
	OTPPolicyPeriod                                           *int                              `json:"otpPolicyPeriod,omitempty"`
	OTPPolicyType                                             *string                           `json:"otpPolicyType,omitempty"`
	OTPSupportedApplications                                  *[]string                         `json:"otpSupportedApplications,omitempty"`
	PasswordPolicy                                            *string                           `json:"passwordPolicy,omitempty"`
	PermanentLockout                                          *bool                             `json:"permanentLockout,omitempty"`
	ProtocolMappers                                           *[]interface{}                    `json:"protocolMappers,omitempty"`
	QuickLoginCheckMilliSeconds                               *int                              `json:"quickLoginCheckMilliSeconds,omitempty"`
	Realm                                                     *string                           `json:"realm,omitempty"`
	RefreshTokenMaxReuse                                      *int                              `json:"refreshTokenMaxReuse,omitempty"`
	RegistrationAllowed                                       *bool                             `json:"registrationAllowed,omitempty"`
	RegistrationEmailAsUsername                               *bool                             `json:"registrationEmailAsUsername,omitempty"`
	RegistrationFlow                                          *string                           `json:"registrationFlow,omitempty"`
	RememberMe                                                *bool                             `json:"rememberMe,omitempty"`
	RequiredActions                                           *[]interface{}                    `json:"requiredActions,omitempty"`
	ResetCredentialsFlow                                      *string                           `json:"resetCredentialsFlow,omitempty"`
	RequiredCredentials                                       *[]string                         `json:"requiredCredentials,omitempty"`
	ResetPasswordAllowed                                      *bool                             `json:"resetPasswordAllowed,omitempty"`
	Roles                                                     *RolesRepresentation              `json:"roles,omitempty"`
	SSOSessionIdleTimeout                                     *int                              `json:"ssoSessionIdleTimeout,omitempty"`
	SSOSessionIdleTimeoutRememberMe                           *int                              `json:"ssoSessionIdleTimeoutRememberMe,omitempty"`
	SSOSessionMaxLifespan                                     *int                              `json:"ssoSessionMaxLifespan,omitempty"`
	SSOSessionMaxLifespanRememberMe                           *int                              `json:"ssoSessionMaxLifespanRememberMe,omitempty"`
	SMTPServer                                                *map[string]string                `json:"smtpServer,omitempty"`
	ScopeMappings                                             *[]interface{}                    `json:"scopeMappings,omitempty"`
	SSLRequired                                               *string                           `json:"sslRequired,omitempty"`
	SupportedLocales                                          *[]string                         `json:"supportedLocales,omitempty"`
	UserFederationMappers                                     *[]interface{}                    `json:"userFederationMappers,omitempty"`
	UserFederationProviders                                   *[]interface{}                    `json:"userFederationProviders,omitempty"`
	UserManagedAccessAllowed                                  *bool                             `json:"userManagedAccessAllowed,omitempty"`
	Users                                                     *[]User                           `json:"users,omitempty"`
	VerifyEmail                                               *bool                             `json:"verifyEmail,omitempty"`
	WebAuthnPolicyAcceptableAaguids                           *[]string                         `json:"webAuthnPolicyAcceptableAaguids,omitempty"`
	WebAuthnPolicyAttestationConveyancePreference             *string                           `json:"webAuthnPolicyAttestationConveyancePreference,omitempty"`
	WebAuthnPolicyAuthenticatorAttachment                     *string                           `json:"webAuthnPolicyAuthenticatorAttachment,omitempty"`
	WebAuthnPolicyAvoidSameAuthenticatorRegister              *bool                             `json:"webAuthnPolicyAvoidSameAuthenticatorRegister,omitempty"`
	WebAuthnPolicyCreateTimeout                               *int                              `json:"webAuthnPolicyCreateTimeout,omitempty"`
	WebAuthnPolicyExtraOrigins                                *[]string                         `json:"webAuthnPolicyExtraOrigins,omitempty"`
	WebAuthnPolicyRequireResidentKey                          *string                           `json:"webAuthnPolicyRequireResidentKey,omitempty"`
	WebAuthnPolicyRpEntityName                                *string                           `json:"webAuthnPolicyRpEntityName,omitempty"`
	WebAuthnPolicyRpId                                        *string                           `json:"webAuthnPolicyRpId,omitempty"`
	WebAuthnPolicySignatureAlgorithms                         *[]string                         `json:"webAuthnPolicySignatureAlgorithms,omitempty"`
	WebAuthnPolicyUserVerificationRequirement                 *string                           `json:"webAuthnPolicyUserVerificationRequirement,omitempty"`
	WebAuthnPolicyPasswordlessAcceptableAaguids               *[]string                         `json:"webAuthnPolicyPasswordlessAcceptableAaguids,omitempty"`
	WebAuthnPolicyPasswordlessAttestationConveyancePreference *string                           `json:"webAuthnPolicyPasswordlessAttestationConveyancePreference,omitempty"`
	WebAuthnPolicyPasswordlessAuthenticatorAttachment         *string                           `json:"webAuthnPolicyPasswordlessAuthenticatorAttachment,omitempty"`
	WebAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister  *bool                             `json:"webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister,omitempty"`
	WebAuthnPolicyPasswordlessCreateTimeout                   *int                              `json:"webAuthnPolicyPasswordlessCreateTimeout,omitempty"`
	WebAuthnPolicyPasswordlessExtraOrigins                    *[]string                         `json:"webAuthnPolicyPasswordlessExtraOrigins,omitempty"`
	WebAuthnPolicyPasswordlessRequireResidentKey              *string                           `json:"webAuthnPolicyPasswordlessRequireResidentKey,omitempty"`
	WebAuthnPolicyPasswordlessRpEntityName                    *string                           `json:"webAuthnPolicyPasswordlessRpEntityName,omitempty"`
	WebAuthnPolicyPasswordlessRpID                            *string                           `json:"webAuthnPolicyPasswordlessRpId,omitempty"`
	WebAuthnPolicyPasswordlessSignatureAlgorithms             *[]string                         `json:"webAuthnPolicyPasswordlessSignatureAlgorithms,omitempty"`
	WebAuthnPolicyPasswordlessUserVerificationRequirement     *string                           `json:"webAuthnPolicyPasswordlessUserVerificationRequirement,omitempty"`
	WaitIncrementSeconds                                      *int                              `json:"waitIncrementSeconds,omitempty"`
}

// AuthenticationFlowRepresentation represents an authentication flow of a realm
type AuthenticationFlowRepresentation struct {
	Alias                    *string                                  `json:"alias,omitempty"`
	AuthenticationExecutions *[]AuthenticationExecutionRepresentation `json:"authenticationExecutions,omitempty"`
	BuiltIn                  *bool                                    `json:"builtIn,omitempty"`
	Description              *string                                  `json:"description,omitempty"`
	ID                       *string                                  `json:"id,omitempty"`
	ProviderID               *string                                  `json:"providerId,omitempty"`
	TopLevel                 *bool                                    `json:"topLevel,omitempty"`
	ProvidedBy               *string                                  `json:"providedBy,omitempty"`
}

// AuthenticationExecutionRepresentation represents the authentication execution of an AuthenticationFlowRepresentation
type AuthenticationExecutionRepresentation struct {
	Authenticator       *string `json:"authenticator,omitempty"`
	AuthenticatorConfig *string `json:"authenticatorConfig,omitempty"`
	AuthenticatorFlow   *bool   `json:"authenticatorFlow,omitempty"`
	AutheticatorFlow    *bool   `json:"autheticatorFlow,omitempty"`
	FlowAlias           *string `json:"flowAlias,omitempty"`
	Priority            *int    `json:"priority,omitempty"`
	Requirement         *string `json:"requirement,omitempty"`
	UserSetupAllowed    *bool   `json:"userSetupAllowed,omitempty"`
}

// CreateAuthenticationExecutionRepresentation contains the provider to be used for a new authentication representation
type CreateAuthenticationExecutionRepresentation struct {
	Provider *string `json:"provider,omitempty"`
}

// CreateAuthenticationExecutionFlowRepresentation contains the provider to be used for a new authentication representation
type CreateAuthenticationExecutionFlowRepresentation struct {
	Alias       *string `json:"alias,omitempty"`
	Description *string `json:"description,omitempty"`
	Provider    *string `json:"provider,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// ModifyAuthenticationExecutionRepresentation is the payload for updating an execution representation
type ModifyAuthenticationExecutionRepresentation struct {
	ID                   *string   `json:"id,omitempty"`
	ProviderID           *string   `json:"providerId,omitempty"`
	AuthenticationConfig *string   `json:"authenticationConfig,omitempty"`
	AuthenticationFlow   *bool     `json:"authenticationFlow,omitempty"`
	Requirement          *string   `json:"requirement,omitempty"`
	FlowID               *string   `json:"flowId"`
	DisplayName          *string   `json:"displayName,omitempty"`
	Alias                *string   `json:"alias,omitempty"`
	RequirementChoices   *[]string `json:"requirementChoices,omitempty"`
	Configurable         *bool     `json:"configurable,omitempty"`
	Level                *int      `json:"level,omitempty"`
	Index                *int      `json:"index,omitempty"`
	Description          *string   `json:"description"`
}

// MultiValuedHashMap represents something
type MultiValuedHashMap struct {
	Empty      *bool    `json:"empty,omitempty"`
	LoadFactor *float32 `json:"loadFactor,omitempty"`
	Threshold  *int32   `json:"threshold,omitempty"`
}

// AuthorizationParameters represents the options to obtain get an authorization
type AuthorizationParameters struct {
	ResponseType *string `json:"code,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	Scope        *string `json:"scope,omitempty"`
	RedirectURI  *string `json:"redirect_uri,omitempty"`
	State        *string `json:"state,omitempty"`
	Nonce        *string `json:"nonce,omitempty"`
	IDTokenHint  *string `json:"id_token_hint,omitempty"`
}

// FormData returns a map of options to be used in SetFormData function
func (p *AuthorizationParameters) FormData() map[string]string {
	m, _ := json.Marshal(p)
	var res map[string]string
	_ = json.Unmarshal(m, &res)
	return res
}

// AuthorizationResponse represents the response to an authorization request.
type AuthorizationResponse struct {
}

// TokenOptions represents the options to obtain a token
type TokenOptions struct {
	ClientID            *string   `json:"client_id,omitempty"`
	ClientSecret        *string   `json:"-"`
	GrantType           *string   `json:"grant_type,omitempty"`
	RefreshToken        *string   `json:"refresh_token,omitempty"`
	Scopes              *[]string `json:"-"`
	Scope               *string   `json:"scope,omitempty"`
	ResponseTypes       *[]string `json:"-"`
	ResponseType        *string   `json:"response_type,omitempty"`
	Permission          *string   `json:"permission,omitempty"`
	Username            *string   `json:"username,omitempty"`
	Password            *string   `json:"password,omitempty"`
	Totp                *string   `json:"totp,omitempty"`
	Code                *string   `json:"code,omitempty"`
	RedirectURI         *string   `json:"redirect_uri,omitempty"`
	ClientAssertionType *string   `json:"client_assertion_type,omitempty"`
	ClientAssertion     *string   `json:"client_assertion,omitempty"`
	SubjectToken        *string   `json:"subject_token,omitempty"`
	RequestedSubject    *string   `json:"requested_subject,omitempty"`
	Audience            *string   `json:"audience,omitempty"`
	RequestedTokenType  *string   `json:"requested_token_type,omitempty"`
}

// FormData returns a map of options to be used in SetFormData function
func (t *TokenOptions) FormData() map[string]string {
	if !NilOrEmptySlice(t.Scopes) {
		t.Scope = StringP(strings.Join(*t.Scopes, " "))
	}
	if !NilOrEmptySlice(t.ResponseTypes) {
		t.ResponseType = StringP(strings.Join(*t.ResponseTypes, " "))
	}
	if NilOrEmpty(t.ResponseType) {
		t.ResponseType = StringP("token")
	}
	m, _ := json.Marshal(t)
	var res map[string]string
	_ = json.Unmarshal(m, &res)
	return res
}

// RequestingPartyTokenOptions represents the options to obtain a requesting party token
type RequestingPartyTokenOptions struct {
	GrantType                     *string   `json:"grant_type,omitempty"`
	Ticket                        *string   `json:"ticket,omitempty"`
	ClaimToken                    *string   `json:"claim_token,omitempty"`
	ClaimTokenFormat              *string   `json:"claim_token_format,omitempty"`
	RPT                           *string   `json:"rpt,omitempty"`
	Permissions                   *[]string `json:"-"`
	PermissionResourceFormat      *string   `json:"permission_resource_format,omitempty"`
	PermissionResourceMatchingURI *bool     `json:"permission_resource_matching_uri,string,omitempty"`
	Audience                      *string   `json:"audience,omitempty"`
	ResponseIncludeResourceName   *bool     `json:"response_include_resource_name,string,omitempty"`
	ResponsePermissionsLimit      *uint32   `json:"response_permissions_limit,omitempty"`
	SubmitRequest                 *bool     `json:"submit_request,string,omitempty"`
	ResponseMode                  *string   `json:"response_mode,omitempty"`
	SubjectToken                  *string   `json:"subject_token,omitempty"`
}

// FormData returns a map of options to be used in SetFormData function
func (t *RequestingPartyTokenOptions) FormData() map[string]string {
	if NilOrEmpty(t.GrantType) { // required grant type for RPT
		t.GrantType = StringP("urn:ietf:params:oauth:grant-type:uma-ticket")
	}
	if t.ResponseIncludeResourceName == nil { // defaults to true if no value set
		t.ResponseIncludeResourceName = BoolP(true)
	}

	m, _ := json.Marshal(t)
	var res map[string]string
	_ = json.Unmarshal(m, &res)
	return res
}

// RequestingPartyPermission is returned by request party token with response type set to "permissions"
type RequestingPartyPermission struct {
	Claims       *map[string]string `json:"claims,omitempty"`
	ResourceID   *string            `json:"rsid,omitempty"`
	ResourceName *string            `json:"rsname,omitempty"`
	Scopes       *[]string          `json:"scopes,omitempty"`
}

// RequestingPartyPermissionDecision is returned by request party token with response type set to "decision"
type RequestingPartyPermissionDecision struct {
	Result *bool `json:"result,omitempty"`
}

// UserSessionRepresentation represents a list of user's sessions
type UserSessionRepresentation struct {
	Clients    *map[string]string `json:"clients,omitempty"`
	ID         *string            `json:"id,omitempty"`
	IPAddress  *string            `json:"ipAddress,omitempty"`
	LastAccess *int64             `json:"lastAccess,omitempty"`
	Start      *int64             `json:"start,omitempty"`
	UserID     *string            `json:"userId,omitempty"`
	Username   *string            `json:"username,omitempty"`
}

// SystemInfoRepresentation represents a system info
type SystemInfoRepresentation struct {
	FileEncoding   *string `json:"fileEncoding,omitempty"`
	JavaHome       *string `json:"javaHome,omitempty"`
	JavaRuntime    *string `json:"javaRuntime,omitempty"`
	JavaVendor     *string `json:"javaVendor,omitempty"`
	JavaVersion    *string `json:"javaVersion,omitempty"`
	JavaVM         *string `json:"javaVm,omitempty"`
	JavaVMVersion  *string `json:"javaVmVersion,omitempty"`
	OSArchitecture *string `json:"osArchitecture,omitempty"`
	OSName         *string `json:"osName,omitempty"`
	OSVersion      *string `json:"osVersion,omitempty"`
	ServerTime     *string `json:"serverTime,omitempty"`
	Uptime         *string `json:"uptime,omitempty"`
	UptimeMillis   *int    `json:"uptimeMillis,omitempty"`
	UserDir        *string `json:"userDir,omitempty"`
	UserLocale     *string `json:"userLocale,omitempty"`
	UserName       *string `json:"userName,omitempty"`
	UserTimezone   *string `json:"userTimezone,omitempty"`
	Version        *string `json:"version,omitempty"`
}

// MemoryInfoRepresentation represents a memory info
type MemoryInfoRepresentation struct {
	Free           *int    `json:"free,omitempty"`
	FreeFormated   *string `json:"freeFormated,omitempty"`
	FreePercentage *int    `json:"freePercentage,omitempty"`
	Total          *int    `json:"total,omitempty"`
	TotalFormated  *string `json:"totalFormated,omitempty"`
	Used           *int    `json:"used,omitempty"`
	UsedFormated   *string `json:"usedFormated,omitempty"`
}

// PasswordPolicy represents the configuration for a supported password policy
type PasswordPolicy struct {
	ConfigType        string `json:"configType,omitempty"`
	DefaultValue      string `json:"defaultValue,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	ID                string `json:"id,omitempty"`
	MultipleSupported bool   `json:"multipleSupported,omitempty"`
}

// ProtocolMapperTypeProperty represents a property of a ProtocolMapperType
type ProtocolMapperTypeProperty struct {
	Name         string         `json:"name,omitempty"`
	Label        string         `json:"label,omitempty"`
	HelpText     string         `json:"helpText,omitempty"`
	Type         string         `json:"type,omitempty"`
	Options      []string       `json:"options,omitempty"`
	DefaultValue EnforcedString `json:"defaultValue,omitempty"`
	Secret       bool           `json:"secret,omitempty"`
	ReadOnly     bool           `json:"readOnly,omitempty"`
}

// ProtocolMapperType represents a type of protocol mapper
type ProtocolMapperType struct {
	ID         string                       `json:"id,omitempty"`
	Name       string                       `json:"name,omitempty"`
	Category   string                       `json:"category,omitempty"`
	HelpText   string                       `json:"helpText,omitempty"`
	Priority   int                          `json:"priority,omitempty"`
	Properties []ProtocolMapperTypeProperty `json:"properties,omitempty"`
}

// ProtocolMapperTypes holds the currently available ProtocolMapperType-s grouped by protocol
type ProtocolMapperTypes struct {
	DockerV2      []ProtocolMapperType `json:"docker-v2,omitempty"`
	SAML          []ProtocolMapperType `json:"saml,omitempty"`
	OpenIDConnect []ProtocolMapperType `json:"openid-connect,omitempty"`
}

// BuiltinProtocolMappers holds the currently available built-in blueprints of ProtocolMapper-s grouped by protocol
type BuiltinProtocolMappers struct {
	SAML          []ProtocolMapperRepresentation `json:"saml,omitempty"`
	OpenIDConnect []ProtocolMapperRepresentation `json:"openid-connect,omitempty"`
}

// ServerInfoRepresentation represents a server info
type ServerInfoRepresentation struct {
	SystemInfo             *SystemInfoRepresentation `json:"systemInfo,omitempty"`
	MemoryInfo             *MemoryInfoRepresentation `json:"memoryInfo,omitempty"`
	PasswordPolicies       []*PasswordPolicy         `json:"passwordPolicies,omitempty"`
	ProtocolMapperTypes    *ProtocolMapperTypes      `json:"protocolMapperTypes,omitempty"`
	BuiltinProtocolMappers *BuiltinProtocolMappers   `json:"builtinProtocolMappers,omitempty"`
	Themes                 *Themes                   `json:"themes,omitempty"`
}

// ThemeRepresentation contains the theme name and locales
type ThemeRepresentation struct {
	Name    string   `json:"name,omitempty"`
	Locales []string `json:"locales,omitempty"`
}

// Themes contains the available keycloak themes with locales
type Themes struct {
	Accounts []ThemeRepresentation `json:"account,omitempty"`
	Admin    []ThemeRepresentation `json:"admin,omitempty"`
	Common   []ThemeRepresentation `json:"common,omitempty"`
	Email    []ThemeRepresentation `json:"email,omitempty"`
	Login    []ThemeRepresentation `json:"login,omitempty"`
	Welcome  []ThemeRepresentation `json:"welcome,omitempty"`
}

// FederatedIdentityRepresentation represents an user federated identity
type FederatedIdentityRepresentation struct {
	IdentityProvider *string `json:"identityProvider,omitempty"`
	UserID           *string `json:"userId,omitempty"`
	UserName         *string `json:"userName,omitempty"`
}

// IdentityProviderRepresentation represents an identity provider
type IdentityProviderRepresentation struct {
	AddReadTokenRoleOnCreate  *bool              `json:"addReadTokenRoleOnCreate,omitempty"`
	Alias                     *string            `json:"alias,omitempty"`
	Config                    *map[string]string `json:"config,omitempty"`
	DisplayName               *string            `json:"displayName,omitempty"`
	Enabled                   *bool              `json:"enabled,omitempty"`
	FirstBrokerLoginFlowAlias *string            `json:"firstBrokerLoginFlowAlias,omitempty"`
	InternalID                *string            `json:"internalId,omitempty"`
	LinkOnly                  *bool              `json:"linkOnly,omitempty"`
	PostBrokerLoginFlowAlias  *string            `json:"postBrokerLoginFlowAlias,omitempty"`
	ProviderID                *string            `json:"providerId,omitempty"`
	StoreToken                *bool              `json:"storeToken,omitempty"`
	TrustEmail                *bool              `json:"trustEmail,omitempty"`
	UpdateProfileFirstLogin   *bool              `json:"updateProfileFirstLogin,omitempty"`
	AuthenticateByDefault     *bool              `json:"authenticateByDefault,omitempty"`
}

// IdentityProviderMapper represents the body of a call to add a mapper to
// an identity provider
type IdentityProviderMapper struct {
	ID                     *string            `json:"id,omitempty"`
	Name                   *string            `json:"name,omitempty"`
	IdentityProviderMapper *string            `json:"identityProviderMapper,omitempty"`
	IdentityProviderAlias  *string            `json:"identityProviderAlias,omitempty"`
	Config                 *map[string]string `json:"config"`
}

// GetResourceParams represents the optional parameters for getting resources
type GetResourceParams struct {
	Deep        *bool   `json:"deep,string,omitempty"`
	First       *int    `json:"first,string,omitempty"`
	Max         *int    `json:"max,string,omitempty"`
	Name        *string `json:"name,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Type        *string `json:"type,omitempty"`
	URI         *string `json:"uri,omitempty"`
	Scope       *string `json:"scope,omitempty"`
	MatchingURI *bool   `json:"matchingUri,string,omitempty"`
	ExactName   *bool   `json:"exactName,string,omitempty"`
}

// GetScopeParams represents the optional parameters for getting scopes
type GetScopeParams struct {
	Deep  *bool   `json:"deep,string,omitempty"`
	First *int    `json:"first,string,omitempty"`
	Max   *int    `json:"max,string,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// GetPolicyParams represents the optional parameters for getting policies
// TODO: more policy params?
type GetPolicyParams struct {
	First      *int    `json:"first,string,omitempty"`
	Max        *int    `json:"max,string,omitempty"`
	Name       *string `json:"name,omitempty"`
	Permission *bool   `json:"permission,string,omitempty"`
	Type       *string `json:"type,omitempty"`
}

// GetPermissionParams represents the optional parameters for getting permissions
type GetPermissionParams struct {
	First    *int    `json:"first,string,omitempty"`
	Max      *int    `json:"max,string,omitempty"`
	Name     *string `json:"name,omitempty"`
	Resource *string `json:"resource,omitempty"`
	Scope    *string `json:"scope,omitempty"`
	Type     *string `json:"type,omitempty"`
}

// GetUsersByRoleParams represents the optional parameters for getting users by role
type GetUsersByRoleParams struct {
	First *int `json:"first,string,omitempty"`
	Max   *int `json:"max,string,omitempty"`
}

// PermissionRepresentation is a representation of a RequestingPartyPermission
type PermissionRepresentation struct {
	DecisionStrategy *DecisionStrategy `json:"decisionStrategy,omitempty"`
	Description      *string           `json:"description,omitempty"`
	ID               *string           `json:"id,omitempty"`
	Logic            *Logic            `json:"logic,omitempty"`
	Name             *string           `json:"name,omitempty"`
	Policies         *[]string         `json:"policies,omitempty"`
	Resources        *[]string         `json:"resources,omitempty"`
	ResourceType     *string           `json:"resourceType,omitempty"`
	Scopes           *[]string         `json:"scopes,omitempty"`
	Type             *string           `json:"type,omitempty"`
}

// CreatePermissionTicketParams represents the optional parameters for getting a permission ticket
type CreatePermissionTicketParams struct {
	ResourceID     *string              `json:"resource_id,omitempty"`
	ResourceScopes *[]string            `json:"resource_scopes,omitempty"`
	Claims         *map[string][]string `json:"claims,omitempty"`
}

// PermissionTicketDescriptionRepresentation represents the parameters returned along with a permission ticket
type PermissionTicketDescriptionRepresentation struct {
	ID                     *string               `json:"id,omitempty"`
	CreatedTimeStamp       *int64                `json:"createdTimestamp,omitempty"`
	UserName               *string               `json:"username,omitempty"`
	Enabled                *bool                 `json:"enabled,omitempty"`
	TOTP                   *bool                 `json:"totp,omitempty"`
	EmailVerified          *bool                 `json:"emailVerified,omitempty"`
	FirstName              *string               `json:"firstName,omitempty"`
	LastName               *string               `json:"lastName,omitempty"`
	Email                  *string               `json:"email,omitempty"`
	DisableCredentialTypes *[]string             `json:"disableCredentialTypes,omitempty"`
	RequiredActions        *[]string             `json:"requiredActions,omitempty"`
	NotBefore              *int64                `json:"notBefore,omitempty"`
	Access                 *AccessRepresentation `json:"access,omitempty"`
}

// AccessRepresentation represents the access parameters returned in the permission ticket description
type AccessRepresentation struct {
	ManageGroupMembership *bool `json:"manageGroupMembership,omitempty"`
	View                  *bool `json:"view,omitempty"`
	MapRoles              *bool `json:"mapRoles,omitempty"`
	Impersonate           *bool `json:"impersonate,omitempty"`
	Manage                *bool `json:"manage,omitempty"`
}

// PermissionTicketResponseRepresentation represents the keycloak response containing the permission ticket
type PermissionTicketResponseRepresentation struct {
	Ticket *string `json:"ticket,omitempty"`
}

// PermissionTicketRepresentation represents the permission ticket contents
type PermissionTicketRepresentation struct {
	AZP         *string                                     `json:"azp,omitempty"`
	Claims      *map[string][]string                        `json:"claims,omitempty"`
	Permissions *[]PermissionTicketPermissionRepresentation `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// PermissionTicketPermissionRepresentation represents the individual permissions in a permission ticket
type PermissionTicketPermissionRepresentation struct {
	Scopes *[]string `json:"scopes,omitempty"`
	RSID   *string   `json:"rsid,omitempty"`
}

// PermissionGrantParams represents the permission which the resource owner is granting to a specific user
type PermissionGrantParams struct {
	ResourceID  *string `json:"resource,omitempty"`
	RequesterID *string `json:"requester,omitempty"`
	Granted     *bool   `json:"granted,omitempty"`
	ScopeName   *string `json:"scopeName,omitempty"`
	TicketID    *string `json:"id,omitempty"`
}

// PermissionGrantResponseRepresentation represents the reply from Keycloack after granting permission
type PermissionGrantResponseRepresentation struct {
	ID          *string `json:"id,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	ResourceID  *string `json:"resource,omitempty"`
	Scope       *string `json:"scope,omitempty"`
	Granted     *bool   `json:"granted,omitempty"`
	RequesterID *string `json:"requester,omitempty"`
}

// GetUserPermissionParams represents the optional parameters for getting user permissions
type GetUserPermissionParams struct {
	ScopeID     *string `json:"scopeId,omitempty"`
	ResourceID  *string `json:"resourceId,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Requester   *string `json:"requester,omitempty"`
	Granted     *bool   `json:"granted,omitempty"`
	ReturnNames *string `json:"returnNames,omitempty"`
	First       *int    `json:"first,string,omitempty"`
	Max         *int    `json:"max,string,omitempty"`
}

// ResourcePolicyRepresentation is a representation of a Policy applied to a resource
type ResourcePolicyRepresentation struct {
	Name             *string           `json:"name,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Scopes           *[]string         `json:"scopes,omitempty"`
	Roles            *[]string         `json:"roles,omitempty"`
	Groups           *[]string         `json:"groups,omitempty"`
	Clients          *[]string         `json:"clients,omitempty"`
	ID               *string           `json:"id,omitempty"`
	Logic            *Logic            `json:"logic,omitempty"`
	DecisionStrategy *DecisionStrategy `json:"decisionStrategy,omitempty"`
	Owner            *string           `json:"owner,omitempty"`
	Type             *string           `json:"type,omitempty"`
	Users            *[]string         `json:"users,omitempty"`
}

// PolicyScopeRepresentation is a representation of a scopes of specific policy
type PolicyScopeRepresentation struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// PolicyResourceRepresentation is a representation of a resource of specific policy
type PolicyResourceRepresentation struct {
	ID   *string `json:"_id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// GetResourcePoliciesParams is a representation of the query params for getting policies
type GetResourcePoliciesParams struct {
	ResourceID *string `json:"resource,omitempty"`
	Name       *string `json:"name,omitempty"`
	Scope      *string `json:"scope,omitempty"`
	First      *int    `json:"first,string,omitempty"`
	Max        *int    `json:"max,string,omitempty"`
}

// GetEventsParams represents the optional parameters for getting events
type GetEventsParams struct {
	Client    *string  `json:"client,omitempty"`
	DateFrom  *string  `json:"dateFrom,omitempty"`
	DateTo    *string  `json:"dateTo,omitempty"`
	First     *int32   `json:"first,string,omitempty"`
	IPAddress *string  `json:"ipAddress,omitempty"`
	Max       *int32   `json:"max,string,omitempty"`
	Type      []string `json:"type,omitempty"`
	UserID    *string  `json:"user,omitempty"`
}

// EventRepresentation is a representation of a Event
type EventRepresentation struct {
	Time      int64             `json:"time,omitempty"`
	Type      *string           `json:"type,omitempty"`
	RealmID   *string           `json:"realmId,omitempty"`
	ClientID  *string           `json:"clientId,omitempty"`
	UserID    *string           `json:"userId,omitempty"`
	SessionID *string           `json:"sessionId,omitempty"`
	IPAddress *string           `json:"ipAddress,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

// CredentialRepresentation is a representations of the credentials
// v7: https://www.keycloak.org/docs-api/7.0/rest-api/index.html#_credentialrepresentation
// v8: https://www.keycloak.org/docs-api/8.0/rest-api/index.html#_credentialrepresentation
type CredentialRepresentation struct {
	// Common part
	CreatedDate *int64  `json:"createdDate,omitempty"`
	Temporary   *bool   `json:"temporary,omitempty"`
	Type        *string `json:"type,omitempty"`
	Value       *string `json:"value,omitempty"`

	// <= v7
	Algorithm         *string             `json:"algorithm,omitempty"`
	Config            *MultiValuedHashMap `json:"config,omitempty"`
	Counter           *int32              `json:"counter,omitempty"`
	Device            *string             `json:"device,omitempty"`
	Digits            *int32              `json:"digits,omitempty"`
	HashIterations    *int32              `json:"hashIterations,omitempty"`
	HashedSaltedValue *string             `json:"hashedSaltedValue,omitempty"`
	Period            *int32              `json:"period,omitempty"`
	Salt              *string             `json:"salt,omitempty"`

	// >= v8
	CredentialData *string `json:"credentialData,omitempty"`
	ID             *string `json:"id,omitempty"`
	Priority       *int32  `json:"priority,omitempty"`
	SecretData     *string `json:"secretData,omitempty"`
	UserLabel      *string `json:"userLabel,omitempty"`
}

// BruteForceStatus is a representation of realm user regarding brute force attack
type BruteForceStatus struct {
	NumFailures   *int    `json:"numFailures,omitempty"`
	Disabled      *bool   `json:"disabled,omitempty"`
	LastIPFailure *string `json:"lastIPFailure,omitempty"`
	LastFailure   *int    `json:"lastFailure,omitempty"`
}

// RequiredActionProviderRepresentation is a representation of required actions
// v15: https://www.keycloak.org/docs-api/15.0/rest-api/index.html#_requiredactionproviderrepresentation
type RequiredActionProviderRepresentation struct {
	Alias         *string            `json:"alias,omitempty"`
	Config        *map[string]string `json:"config,omitempty"`
	DefaultAction *bool              `json:"defaultAction,omitempty"`
	Enabled       *bool              `json:"enabled,omitempty"`
	Name          *string            `json:"name,omitempty"`
	Priority      *int32             `json:"priority,omitempty"`
	ProviderID    *string            `json:"providerId,omitempty"`
}

type UnregisteredRequiredActionProviderRepresentation struct {
	Name       *string `json:"name,omitempty"`
	ProviderID *string `json:"providerId,omitempty"`
}

// ManagementPermissionRepresentation is a representation of management permissions
// v18: https://www.keycloak.org/docs-api/18.0/rest-api/#_managementpermissionreference
type ManagementPermissionRepresentation struct {
	Enabled          *bool              `json:"enabled,omitempty"`
	Resource         *string            `json:"resource,omitempty"`
	ScopePermissions *map[string]string `json:"scopePermissions,omitempty"`
}

// GetClientUserSessionsParams represents the optional parameters for getting user sessions associated with the client
type GetClientUserSessionsParams struct {
	First *int `json:"first,string,omitempty"`
	Max   *int `json:"max,string,omitempty"`
}

// OrganizationInviteUserParams represents the parameters for inviting a new user to an organization
type OrganizationInviteUserParams struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

// FormData returns form data for a given OrganizationInviteUserParams
func (v *OrganizationInviteUserParams) FormData() map[string]string {
	m, _ := json.Marshal(v)
	var res map[string]string
	_ = json.Unmarshal(m, &res)
	return res
}

// GetMembersParams represents the optional parameters for getting members of an organization
type GetMembersParams struct {
	Exact          *bool           `json:"exact,string,omitempty"`
	First          *int            `json:"first,string,omitempty"`
	Max            *int            `json:"max,string,omitempty"`
	MembershipType *MembershipType `json:"membershipetype,omitempty"`
	Search         *string         `json:"search,omitempty"`
}

// MembershipType represent the membership type of an organization member.
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#MembershipType
type MembershipType struct{}

// MemberRepresentation represents a member of an organization
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#MemberRepresentation
type MemberRepresentation struct {
	User
	// Type not defined in the Keycloak doc so I left it unexported. Help if you have more information
	MembershipType *MembershipType `json:"membershipetype,omitempty"`
}

// GetOrganizationsParams represents the optional parameters for getting organizations
type GetOrganizationsParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"`
	Exact               *bool   `json:"exact,string,omitempty"`
	First               *int    `json:"first,string,omitempty"`
	Max                 *int    `json:"max,string,omitempty"`
	Q                   *string `json:"q,omitempty"`
	Search              *string `json:"search,omitempty"`
}

// OrganizationDomainRepresentation is a representation of an organization's domain
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#OrganizationDomainRepresentation
type OrganizationDomainRepresentation struct {
	Name     *string `json:"name,omitempty"`
	Verified *bool   `json:"verified,omitempty"`
}

// OrganizationRepresentation is a representation of an organization
// v26: https://www.keycloak.org/docs-api/latest/rest-api/index.html#OrganizationRepresentation
type OrganizationRepresentation struct {
	ID                *string                             `json:"id,omitempty"`
	Name              *string                             `json:"name,omitempty"`
	Alias             *string                             `json:"alias,omitempty"`
	Enabled           *bool                               `json:"enabled,omitempty"`
	Description       *string                             `json:"description,omitempty"`
	RedirectURL       *string                             `json:"redirectUrl,omitempty"`
	Attributes        *map[string][]string                `json:"attributes,omitempty"`
	Domains           *[]OrganizationDomainRepresentation `json:"domains,omitempty"`
	Members           *[]MemberRepresentation             `json:"members,omitempty"`
	IdentityProviders *[]IdentityProviderRepresentation   `json:"identityProviders,omitempty"`
}

// prettyStringStruct returns struct formatted into pretty string
func prettyStringStruct(t interface{}) string {
	json, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		return ""
	}

	return string(json)
}

// Stringer implementations for all struct types
func (v *CertResponseKey) String() string                           { return prettyStringStruct(v) }
func (v *CertResponse) String() string                              { return prettyStringStruct(v) }
func (v *IssuerResponse) String() string                            { return prettyStringStruct(v) }
func (v *ResourcePermission) String() string                        { return prettyStringStruct(v) }
func (v *PermissionResource) String() string                        { return prettyStringStruct(v) }
func (v *PermissionScope) String() string                           { return prettyStringStruct(v) }
func (v *IntroSpectTokenResult) String() string                     { return prettyStringStruct(v) }
func (v *User) String() string                                      { return prettyStringStruct(v) }
func (v *SetPasswordRequest) String() string                        { return prettyStringStruct(v) }
func (v *Component) String() string                                 { return prettyStringStruct(v) }
func (v *KeyStoreConfig) String() string                            { return prettyStringStruct(v) }
func (v *ActiveKeys) String() string                                { return prettyStringStruct(v) }
func (v *Key) String() string                                       { return prettyStringStruct(v) }
func (v *Attributes) String() string                                { return prettyStringStruct(v) }
func (v *Access) String() string                                    { return prettyStringStruct(v) }
func (v *UserGroup) String() string                                 { return prettyStringStruct(v) }
func (v *GetUsersParams) String() string                            { return prettyStringStruct(v) }
func (v *GetComponentsParams) String() string                       { return prettyStringStruct(v) }
func (v *ExecuteActionsEmail) String() string                       { return prettyStringStruct(v) }
func (v *Group) String() string                                     { return prettyStringStruct(v) }
func (v *GroupsCount) String() string                               { return prettyStringStruct(v) }
func (obj *GetGroupsParams) String() string                         { return prettyStringStruct(obj) }
func (v *CompositesRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *Role) String() string                                      { return prettyStringStruct(v) }
func (v *GetRoleParams) String() string                             { return prettyStringStruct(v) }
func (v *ClientMappingsRepresentation) String() string              { return prettyStringStruct(v) }
func (v *MappingsRepresentation) String() string                    { return prettyStringStruct(v) }
func (v *ClientScope) String() string                               { return prettyStringStruct(v) }
func (v *ClientScopeAttributes) String() string                     { return prettyStringStruct(v) }
func (v *ProtocolMappers) String() string                           { return prettyStringStruct(v) }
func (v *ProtocolMappersConfig) String() string                     { return prettyStringStruct(v) }
func (v *Client) String() string                                    { return prettyStringStruct(v) }
func (v *ResourceServerRepresentation) String() string              { return prettyStringStruct(v) }
func (v *RoleDefinition) String() string                            { return prettyStringStruct(v) }
func (v *AbstractPolicyRepresentation) String() string              { return prettyStringStruct(v) }
func (v *PolicyRepresentation) String() string                      { return prettyStringStruct(v) }
func (v *RolePolicyRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *JSPolicyRepresentation) String() string                    { return prettyStringStruct(v) }
func (v *ClientPolicyRepresentation) String() string                { return prettyStringStruct(v) }
func (v *TimePolicyRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *UserPolicyRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *AggregatedPolicyRepresentation) String() string            { return prettyStringStruct(v) }
func (v *GroupPolicyRepresentation) String() string                 { return prettyStringStruct(v) }
func (v *GroupDefinition) String() string                           { return prettyStringStruct(v) }
func (v *ResourceRepresentation) String() string                    { return prettyStringStruct(v) }
func (v *ResourceOwnerRepresentation) String() string               { return prettyStringStruct(v) }
func (v *ScopeRepresentation) String() string                       { return prettyStringStruct(v) }
func (v *ProtocolMapperRepresentation) String() string              { return prettyStringStruct(v) }
func (v *GetClientsParams) String() string                          { return prettyStringStruct(v) }
func (v *UserInfoAddress) String() string                           { return prettyStringStruct(v) }
func (v *UserInfo) String() string                                  { return prettyStringStruct(v) }
func (v *RolesRepresentation) String() string                       { return prettyStringStruct(v) }
func (v *RealmRepresentation) String() string                       { return prettyStringStruct(v) }
func (v *MultiValuedHashMap) String() string                        { return prettyStringStruct(v) }
func (t *TokenOptions) String() string                              { return prettyStringStruct(t) }
func (t *RequestingPartyTokenOptions) String() string               { return prettyStringStruct(t) }
func (v *RequestingPartyPermission) String() string                 { return prettyStringStruct(v) }
func (v *UserSessionRepresentation) String() string                 { return prettyStringStruct(v) }
func (v *SystemInfoRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *MemoryInfoRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *ServerInfoRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *FederatedIdentityRepresentation) String() string           { return prettyStringStruct(v) }
func (v *IdentityProviderRepresentation) String() string            { return prettyStringStruct(v) }
func (v *GetResourceParams) String() string                         { return prettyStringStruct(v) }
func (v *GetScopeParams) String() string                            { return prettyStringStruct(v) }
func (v *GetPolicyParams) String() string                           { return prettyStringStruct(v) }
func (v *GetPermissionParams) String() string                       { return prettyStringStruct(v) }
func (v *GetUsersByRoleParams) String() string                      { return prettyStringStruct(v) }
func (v *PermissionRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *CreatePermissionTicketParams) String() string              { return prettyStringStruct(v) }
func (v *PermissionTicketDescriptionRepresentation) String() string { return prettyStringStruct(v) }
func (v *AccessRepresentation) String() string                      { return prettyStringStruct(v) }
func (v *PermissionTicketResponseRepresentation) String() string    { return prettyStringStruct(v) }
func (v *PermissionTicketRepresentation) String() string            { return prettyStringStruct(v) }
func (v *PermissionTicketPermissionRepresentation) String() string  { return prettyStringStruct(v) }
func (v *PermissionGrantParams) String() string                     { return prettyStringStruct(v) }
func (v *PermissionGrantResponseRepresentation) String() string     { return prettyStringStruct(v) }
func (v *GetUserPermissionParams) String() string                   { return prettyStringStruct(v) }
func (v *ResourcePolicyRepresentation) String() string              { return prettyStringStruct(v) }
func (v *GetResourcePoliciesParams) String() string                 { return prettyStringStruct(v) }
func (v *CredentialRepresentation) String() string                  { return prettyStringStruct(v) }
func (v *RequiredActionProviderRepresentation) String() string      { return prettyStringStruct(v) }
func (v *BruteForceStatus) String() string                          { return prettyStringStruct(v) }
func (v *GetClientUserSessionsParams) String() string               { return prettyStringStruct(v) }
func (v *GetOrganizationsParams) String() string                    { return prettyStringStruct(v) }
func (v *OrganizationInviteUserParams) String() string              { return prettyStringStruct(v) }
func (v *GetMembersParams) String() string                          { return prettyStringStruct(v) }
func (v *MembershipType) String() string                            { return prettyStringStruct(v) }
func (v *MemberRepresentation) String() string                      { return prettyStringStruct(v) }
func (v *OrganizationDomainRepresentation) String() string          { return prettyStringStruct(v) }
func (v *OrganizationRepresentation) String() string                { return prettyStringStruct(v) }
