package common

import lib "github.com/getbrevo/brevo-go/lib"

// ProviderData is passed to all resources and data sources via Configure.
type ProviderData struct {
	Client *lib.APIClient
	APIKey string
}
