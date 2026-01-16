package messages

// RenderAlert renders a footer with just a message
func RenderAlert(props Props) string {
	content := ColorizeKeys(props, " "+props.Message)
	return props.Style.Footer.Width(props.Width).Render(content)
}
