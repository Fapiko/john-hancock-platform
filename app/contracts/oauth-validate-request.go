package contracts

type OAuthValidateRequest struct {
	AccessToken string `json:"accessToken"`
	Provider    string `json:"provider"`
}
