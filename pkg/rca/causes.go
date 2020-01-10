package rca

type Cause string

const (
	CauseBootstrapTimeout Cause = "Timeout waiting for cluster to bootstrap"
	CauseClusterTimeout   Cause = "Timeout waiting for cluster to initialize"
)

func (c Cause) IsInfra() bool {
	switch c {
	case CauseBootstrapTimeout, CauseClusterTimeout:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer
func (c Cause) String() string {
	return string(c)
}
