package footer

// renderMessage renders a footer with just a message
func renderMessage(props Props) string {
	return props.Styles.Footer.Width(props.Width).Render(" " + props.Message)
}
