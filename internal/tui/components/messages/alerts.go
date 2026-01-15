package messages

// RenderAlert renders a footer with just a message
func RenderAlert(props Props) string {
	return props.Style.Footer.Width(props.Width).Render(" " + props.Message)
}
