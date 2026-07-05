package domains

type DockerLogLine struct {
	ContainerID string `json:"containerId"`
	Stream      string `json:"stream"`
	Line        string `json:"line"`
	Timestamp   int64  `json:"timestamp"`
}

type DockerLogResult struct {
	ContainerID string          `json:"containerId"`
	Lines       []DockerLogLine `json:"lines"`
}
