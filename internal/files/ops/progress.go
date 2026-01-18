package ops

import (
	"io"
	"time"

	"github.com/zulfikawr/fm/internal/files/core"
)

type progressWriter struct {
	io.Writer
	Total    int64
	Current  int64
	Label    string
	progChan chan<- core.Progress
	lastUpd  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	if pw.Current == 0 && pw.progChan != nil {
		select {
		case pw.progChan <- core.Progress{
			Percent: 0,
			Label:   pw.Label + "...",
		}:
		default:
		}
	}
	n, err := pw.Writer.Write(p)
	pw.Current += int64(n)
	if pw.progChan != nil && (time.Since(pw.lastUpd) > 100*time.Millisecond || pw.Current == pw.Total) {
		pw.lastUpd = time.Now()
		percent := 0.0
		if pw.Total > 0 {
			percent = float64(pw.Current) / float64(pw.Total)
		} else if pw.Current > 0 {
			percent = 1.0
		}
		select {
		case pw.progChan <- core.Progress{
			Percent: percent,
			Label:   pw.Label + "...",
		}:
		default:
		}
	}
	if pw.Current == pw.Total && pw.progChan != nil {
		time.Sleep(100 * time.Millisecond)
	}
	return n, err
}
