package contracts

import "time"

type CreateCertificateRequest struct {
	Name                    string    `json:"name"`
	KeyId                   string    `json:"keyId"`
	KeyPassword             string    `json:"keyPassword"`
	KeyUsages               []string  `json:"keyUsages"`
	CommonName              string    `json:"commonName"`
	SubjectAlternativeNames []string  `json:"subjectAlternativeNames"`
	Expiration              time.Time `json:"expiration"`
	CAKeyPassword           string    `json:"caKeyPassword"`
}
