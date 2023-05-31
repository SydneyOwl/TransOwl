package terminal

type Terminal struct {
	Device  Device `json:"device"`
	User    User   `json:"user"`
	FoundAt string `json:"-"`
}
