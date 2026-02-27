package views

func sourceLabel(source string) string {
	switch source {
	case "local":
		return "Local"
	case "make":
		return "Makefile"
	default:
		return "User"
	}
}
