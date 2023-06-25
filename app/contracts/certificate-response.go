package contracts

import (
	"crypto/x509/pkix"
	"time"
)

type CertificateResponse struct {
	ID                 string    `json:"id"`
	OwnerID            string    `json:"ownerId"`
	Name               string    `json:"name"`
	Type               string    `json:"type"`
	Created            time.Time `json:"created"`
	KeyID              string    `json:"keyId"`
	SignatureAlgorithm string    `json:"signatureAlgorithm"`
	PublicKeyAlgorithm string    `json:"publicKeyAlgorithm"`
	Version            int       `json:"version"`
	SerialNumber       int       `json:"serialNumber"`
	Issuer             *PkixName `json:"issuer"`
	Subject            *PkixName `json:"subject"`
	NotBefore          time.Time `json:"notBefore"`
	NotAfter           time.Time `json:"notAfter"`
	KeyUsage           []string  `json:"keyUsage"`
	ExtKeyUsage        []string  `json:"extKeyUsage"`
	IsCA               bool      `json:"isCA"`
	MaxPathLen         int       `json:"maxPathLen"`
	MaxPathLenZero     bool      `json:"maxPathLenZero"`
	DNSNames           []string  `json:"sanDNSNames"`
}

type PkixName struct {
	Country            []string `json:"country"`
	Organization       []string `json:"organization"`
	OrganizationalUnit []string `json:"organizationalUnit"`
	Locality           []string `json:"locality"`
	Province           []string `json:"province"`
	StreetAddress      []string `json:"streetAddress"`
	PostalCode         []string `json:"postalCode"`
	SerialNumber       string   `json:"serialNumber"`
	CommonName         string   `json:"commonName"`
}

func (n *PkixName) FromName(name *pkix.Name) {
	n.Country = name.Country
	n.Organization = name.Organization
	n.OrganizationalUnit = name.OrganizationalUnit
	n.Locality = name.Locality
	n.Province = name.Province
	n.StreetAddress = name.StreetAddress
	n.PostalCode = name.PostalCode
	n.SerialNumber = name.SerialNumber
	n.CommonName = name.CommonName
}
