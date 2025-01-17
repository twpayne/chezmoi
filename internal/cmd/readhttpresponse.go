package cmd

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

const httpProgressWidth = 38

type httpProgressModel struct {
	url           string
	contentLength int
	progress      progress.Model
	canceled      bool
}

type httpSpinnerModel struct {
	url      string
	spinner  spinner.Model
	canceled bool
}

type bytesReadMsg int

type doneMsg struct {
	err error
}

func (m httpProgressModel) Canceled() bool {
	return m.canceled
}

func (m httpProgressModel) Init() tea.Cmd {
	return nil
}

func (m httpProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bytesReadMsg:
		cmd := m.progress.SetPercent(float64(msg) / float64(m.contentLength))
		return m, cmd
	case doneMsg:
		return m, tea.Quit
	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model.(progress.Model) //nolint:forcetypeassert
		return m, cmd
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

func (m httpProgressModel) View() string {
	return "[" + m.progress.View() + "] " + m.url
}

func (m httpSpinnerModel) Canceled() bool {
	return m.canceled
}

func (m httpSpinnerModel) Init() tea.Cmd {
	return nil
}

func (m httpSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bytesReadMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(m.spinner.Tick())
		return m, cmd
	case doneMsg:
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

func (m httpSpinnerModel) View() string {
	return "[" + m.spinner.View() + "] " + m.url
}

func (c *Config) readHTTPResponse(url string, resp *http.Response) ([]byte, error) {
	switch {
	case c.noTTY || !c.Progress.Value(c.progressAutoFunc):
		return io.ReadAll(resp.Body)

	case resp.ContentLength >= 0:
		httpProgress := progress.New(
			progress.WithWidth(httpProgressWidth),
		)
		httpProgress.Full = '#'
		httpProgress.FullColor = ""
		httpProgress.Empty = ' '
		httpProgress.EmptyColor = ""
		httpProgress.ShowPercentage = false

		model := httpProgressModel{
			url:           url,
			contentLength: int(resp.ContentLength),
			progress:      httpProgress,
		}

		return runReadHTTPResponse(model, resp)

	default:
		httpSpinner := spinner.New(
			spinner.WithSpinner(spinner.Spinner{
				Frames: makeNightriderFrames("+", ' ', httpProgressWidth),
				FPS:    time.Second / 60,
			}),
		)

		model := httpSpinnerModel{
			url:     resp.Request.URL.String(),
			spinner: httpSpinner,
		}

		return runReadHTTPResponse(model, resp)
	}
}

type hookWriter struct {
	onWrite func([]byte) (int, error)
}

func (w *hookWriter) Write(p []byte) (int, error) {
	return w.onWrite(p)
}

func runReadHTTPResponse(model cancelableModel, resp *http.Response) ([]byte, error) {
	program := tea.NewProgram(model)

	bytesRead := 0
	hookWriter := &hookWriter{
		onWrite: func(p []byte) (int, error) {
			bytesRead += len(p)
			program.Send(bytesReadMsg(bytesRead))
			return len(p), nil
		},
	}

	var data []byte
	var err error
	go func() {
		data, err = io.ReadAll(io.TeeReader(resp.Body, hookWriter))
		program.Send(doneMsg{
			err: err,
		})
	}()

	if model, err := program.Run(); err != nil {
		return nil, err
	} else if model.(cancelableModel).Canceled() { //nolint:forcetypeassert
		return nil, chezmoi.ExitCodeError(0)
	}

	return data, err
}

func makeNightriderFrames(shape string, padding rune, width int) []string {
	delta := width - len(shape)
	if delta <= 0 {
		return []string{shape[:width]}
	}
	frames := make([]string, 2*delta)
	paddingStr := string([]rune{padding})
	for i := 0; i <= delta; i++ {
		frames[i] = strings.Repeat(paddingStr, i) + shape + strings.Repeat(paddingStr, delta-i)
	}
	for i := delta + 1; i < 2*delta; i++ {
		frames[i] = frames[2*delta-i]
	}
	return frames
}
