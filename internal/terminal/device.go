package terminal

type Device struct {
	OS   string `json:"os,omitempty"`
	Arch string `json:"arch,omitempty"`
}
