package oidcConnector

import (
	model "../model"
)

var OIDConnection model.OIDConnection

// func provider() {
// 	ctx := context.Background()
// 	provider, err := oidc.NewProvider(ctx, OIDConnection.DiscoveryEndpoint)
// 	if err != nil {
// 		// handle error
// 	}
//
// 	// Configure an OpenID Connect aware OAuth2 client.
// 	oauth2Config := oauth2.Config{
// 		ClientID:     OIDConnection.ClientID,
// 		ClientSecret: OIDConnection.ClientSecret,
// 		RedirectURL:  OIDConnection.RedirectURL,
//
// 		// Discovery returns the OAuth2 endpoints.
// 		Endpoint: provider.Endpoint(),
//
// 		// "openid" is a required scope for OpenID Connect flows.
// 		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
// 	}
// }
