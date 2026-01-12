package state

// ProgressState holds progress bar state
type ProgressState struct {
	Visible bool
	Percent float64
	Label   string
}

// Show shows the progress bar with a label
func (p *ProgressState) Show(label string) {
	p.Visible = true
	p.Label = label
	p.Percent = 0
}

// Hide hides the progress bar
func (p *ProgressState) Hide() {
	p.Visible = false
	p.Percent = 0
	p.Label = ""
}

// Update updates the progress percentage
func (p *ProgressState) Update(percent float64) {
	p.Percent = percent
}
