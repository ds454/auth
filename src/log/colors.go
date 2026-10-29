package log

const (
	colorReset  = "\033[0m"
	colorGray   = "\033[90m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
)

func levelColor(level string) string {
	switch level {
	case "ERROR":
		return colorRed
	case "WARN":
		return colorYellow
	case "INFO":
		return colorGreen
	case "DEBUG":
		return colorBlue
	default:
		return colorReset
	}
}
