package messages

type PageType int

const (
	LoadingPage PageType = iota
	ErrorPage
)

// String returns the string representation of a PageType
func (p PageType) String() string {
	switch p {
	case LoadingPage:
		return "loading"
	case ErrorPage:
		return "error"
	default:
		return "unknown"
	}
}

type NavigateMsg struct {
	Page PageType
	Data interface{} // Optional data to pass to the new page
}

type ErrorMsg struct {
	Err     error
	Message string
	Fatal   bool // If true, the application should exit
}

type HealthCheckResult struct {
	Success   bool
	Status    string
	Latency   string
	Error     error
	Timestamp string
}

type HealthCheckMsg HealthCheckResult
