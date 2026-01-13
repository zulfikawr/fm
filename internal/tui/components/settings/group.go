package settings

// renderGroups renders the settings groups into a slice of rows
func renderGroups(props Props) []string {
	var rows []string
	rows = append(rows, "") // Add a line above the first group

	currentIndex := 0
	for i, g := range props.Groups {
		if i > 0 {
			rows = append(rows, "")
		}

		rows = append(rows, props.Styles.SettingsHeader.Width(props.Width).Render(g.Title))
		for _, sItem := range g.Settings {
			rows = append(rows, renderSettingRow(props, sItem, currentIndex == props.Cursor))
			currentIndex++
		}
	}
	return rows
}
