package internal

type Certificate struct {
	Chain []byte `json:"chain"`
	Key   []byte `json:"key"`
	Host  string `json:"host"`
}
