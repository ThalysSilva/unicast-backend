package models

type Secrets struct {
	AccessToken  []byte
	RefreshToken []byte
	Jwe          []byte
}
