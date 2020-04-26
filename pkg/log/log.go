package log

import (
	"fmt"
	"io"
	"time"

	spinner "github.com/briandowns/spinner"
)

// Print displays and returns a new spinner
func Print(title string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[37], 100*time.Millisecond)
	_ = s.Color("fgHiCyan")
	s.Prefix = " "
	s.Suffix = fmt.Sprintf(" %s", title)
	s.FinalMSG = fmt.Sprintf("✔ %s complete\n", title)
	s.Start()
	return s
}

// Fprint displays to the passed io.Writer and returns a new spinner
func Fprint(out io.Writer, title string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[37], 100*time.Millisecond)
	_ = s.Color("fgHiCyan")
	s.Writer = out
	s.Prefix = " "
	s.Suffix = fmt.Sprintf(" %s ", title)
	s.FinalMSG = fmt.Sprintf("✔ %s complete\n", title)
	s.Start()
	return s
}
