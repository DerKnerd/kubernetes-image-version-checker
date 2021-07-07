package messaging

type Message struct {
	UsedVersion   string
	LatestVersion string
	Image         string
	ParentName    string
	EntityType    string
	Cpu           int
}
