package context

// GetPromptText returns the unstyled prompt text for a given input mode
func GetPromptText(mode InputMode, altMode bool) string {
	switch mode {
	case InputGoto:
		label := "Local"
		if altMode {
			label = "Remote"
		}
		return "Go to (" + label + "): "
	case InputAuth:
		label := "Password"
		if altMode {
			label = "PEM Path"
		}
		return label + ": "
	case InputSearch:
		return "Filter: "
	case InputFuzzySearch:
		return "Search: "
	case InputRename:
		return "Rename: "
	case InputZip:
		return "Zip name: "
	case InputUnzip:
		return "Unzip to: "
	case InputCreate:
		label := "File"
		if altMode {
			label = "Folder"
		}
		return "Create (" + label + "): "
	case InputConflictRename:
		return "New name: "
	case InputKeybinding:
		return "Bind: "
	default:
		return ""
	}
}
