package contracts

type CreateKeyRequest struct {
	Algorithm KeyAlgorithm `json:"algorithm"`
	Name      string       `json:"name"`
	Password  string       `json:"password"`
}
