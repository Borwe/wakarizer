package main

import (
	"fmt"
	"os"
	"strings"

	"wakarizer/wakatime"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wakatime/wakatime-cli/pkg/log"
	ini "gopkg.in/ini.v1"
)

type Status int
type Msg[T any] struct{ value T }

const (
	STARTING Status = iota
	GOTTEN_HOME_DIR
	ASKING_API_KEY
	SETUP_ASKING_LANGUAGE_TYPES
	ASKING_LANGUAGE_TYPES
	VALIDATE_GETTING_LANGUAGES
	START_MAIN_ACTIVITY
	INSIDE_MAIN_ACTIVITY
)

type LanguagesInfo struct {
	wakatime_cfg   string
	languages      []string
	language_index int
}

func NewLanguagesInfo() LanguagesInfo {
	return LanguagesInfo{
		language_index: 0,
	}
}

func (l *LanguagesInfo) getKey() string {
	cfg, err := ini.Load(l.wakatime_cfg)
	if err != nil {
		fmt.Printf("Couldn't load cfg file from %s\n", l.wakatime_cfg)
		os.Exit(1)
	}

	return cfg.Section("settings").Key("api_key").String()
}

func (l *LanguagesInfo) setLanguages(languages []string) {
	l.languages = languages
	l.language_index = 0
}

func (l *LanguagesInfo) updateLanguageIndex() {
	max := len(l.languages)
	l.language_index += 1
	if l.language_index >= max {
		l.language_index = 0
	}
}

// Perform the actual wakatime hearbeat
func (l *LanguagesInfo) doHeartBeat() {
	key := l.getKey()
	_, w, error := os.Pipe()
	if error != nil {
		fmt.Fprintln(os.Stderr, "Error creating Pipe")
		os.Exit(1)
	}
	log.SetOutput(w)
	for {
		wakatime.Execute(l.languages[l.language_index], key)
		l.updateLanguageIndex()
	}
}

// Returns true if api-key field is set
// otherwise returns false
func (l *LanguagesInfo) checkIfApiKeySet() tea.Msg {
	cfg, err := ini.Load(l.wakatime_cfg)
	if err != nil {
		//probably file doesn't exist so create it
		f, err := os.OpenFile(l.wakatime_cfg, os.O_CREATE, 0777)
		f.Close()
		cfg, err = ini.Load(l.wakatime_cfg)
		if err != nil {
			fmt.Printf("Error occured with loading config file: %s, with error: %s", l.wakatime_cfg, err)
			return tea.Quit
		}
	}

	if !cfg.HasSection("settings") {
		cfg.NewSection("settings")
		cfg.SaveTo(l.wakatime_cfg)
	}
	settings := cfg.Section("settings")
	var result Msg[bool]
	if !settings.HasKey("api_key") {
		result.value = false
	} else {
		result.value = true
	}
	return result
}

func (l *LanguagesInfo) checkIfLangsInputEmpty() tea.Msg {
	var result Msg[bool]
	for _, v := range l.languages {
		if len(v) == 0 {
			result.value = false
			return result
		}
	}
	result.value = true
	return result
}

func (l *LanguagesInfo) setApiKey(m model) {
	cfg, err := ini.Load(l.wakatime_cfg)
	if err != nil {
		fmt.Printf("Error, %s occured trying to open %s", err, l.wakatime_cfg)
		os.Exit(1)
	}

	cfg.Section("settings").NewKey("api_key", m.text_input.Value())
	if cfg.SaveTo(l.wakatime_cfg) != nil {
		fmt.Printf("Failed to save %s file", l.wakatime_cfg)
		os.Exit(1)
	}
}

var LangInfo LanguagesInfo

func init() {
	LangInfo = NewLanguagesInfo()
}

type model struct {
	text_input textinput.Model
	state      Status
	spinner    spinner.Model
}

func newModel() model {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#11EE11"))

	return model{state: STARTING,
		text_input: textinput.New(),
		spinner:    spin,
	}
}

func getHomeDir() tea.Msg {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Couldn't get home dir of your os")
		return tea.Quit
	}

	var result Msg[string]
	result.value = home + string(os.PathSeparator) + ".wakarizer.cfg"
	return result
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	//handle key presses
	case tea.KeyMsg:
		{
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEnter:
				{
					switch m.state {
					case ASKING_API_KEY:
						{
							m.state = SETUP_ASKING_LANGUAGE_TYPES
							LangInfo.setApiKey(m)
							return m, textinput.Blink
						}
					case ASKING_LANGUAGE_TYPES:
						{
							m.state = VALIDATE_GETTING_LANGUAGES
							LangInfo.setLanguages(strings.Split(m.text_input.Value(), " "))
							return m, LangInfo.checkIfLangsInputEmpty
						}
					}
				}
			case tea.KeyCtrlV:
				{
					if m.state == ASKING_API_KEY || m.state == ASKING_LANGUAGE_TYPES {
						return m, textinput.Paste
					}
				}
			}
		}

	case Msg[string]:
		{
			switch m.state {
			case STARTING:
				{
					m.state = GOTTEN_HOME_DIR
					LangInfo.wakatime_cfg = msg.value
					return m, LangInfo.checkIfApiKeySet
				}
			}
		}

	case Msg[bool]:
		{
			switch m.state {
			case GOTTEN_HOME_DIR:
				{
					if msg.value == true {
						m.state = SETUP_ASKING_LANGUAGE_TYPES
						return m, textinput.Blink
					} else {
						m.state = ASKING_API_KEY
						m.text_input.EchoMode = textinput.EchoPassword
						m.text_input.Placeholder = "Enter Wakatime Api Key"
						m.text_input.Width = 100
						m.text_input.CharLimit = 100
						m.text_input.Focus()
						return m, textinput.Blink
					}
				}
			case VALIDATE_GETTING_LANGUAGES:
				{
					if msg.value == true {
						m.state = START_MAIN_ACTIVITY
						return m, m.spinner.Tick
					} else {
						m.state = ASKING_LANGUAGE_TYPES
						return m, textinput.Blink
					}
				}
			}
		}
	}

	switch m.state {
	//do this when starting
	case STARTING:
		return m, getHomeDir
	case ASKING_API_KEY, ASKING_LANGUAGE_TYPES:
		{
			input, cmd := m.text_input.Update(msg)
			m.text_input = input
			return m, cmd
		}
	case SETUP_ASKING_LANGUAGE_TYPES:
		{
			m.state = ASKING_LANGUAGE_TYPES
			m.text_input = textinput.New()
			m.text_input.Placeholder = "rs ts java...example of how to enter file extentions"
			m.text_input.Width = 100
			m.text_input.Focus()
			return m, textinput.Blink
		}
	case START_MAIN_ACTIVITY:
		{
			go LangInfo.doHeartBeat()
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			m.state = INSIDE_MAIN_ACTIVITY
			return m, cmd
		}
	case INSIDE_MAIN_ACTIVITY:
		{
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) showMainTitle() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#EE1111")).
		Border(lipgloss.DoubleBorder(), true).
		Render("\nWelcome to Wakarizer\n by @Borwe brian.orwe@gmail.com\n")
}

func (m model) showTitle(title string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#EEEEEE")).
		Render("\n" + title)
}

func (m model) showFooter(title string) string {
	return lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.Color("#EE1111")).
		Render(title)
}

func (m model) View() string {
	string_return := m.showMainTitle()

	switch m.state {
	case STARTING:
		string_return += m.showTitle("\nLoading....")
	case GOTTEN_HOME_DIR:
		string_return += m.showTitle(fmt.Sprintf("\nGotten config @ %s", LangInfo.wakatime_cfg))
	case ASKING_API_KEY:
		{
			string_return += fmt.Sprintf("%s\n\n%s", m.showTitle("\nWhat is your api key?"), m.text_input.View())
		}
	case ASKING_LANGUAGE_TYPES:
		string_return += fmt.Sprintf("%s\n\n%s",
			m.showTitle("\nWhat Languages You want to wakarize, give atleast 1? \n"+
				"(write the extention, and seperate with spaces for others\n"+
				"eg: rs ts java\n"+
				"This is for rust, typescript and java)"), m.text_input.View())

	case START_MAIN_ACTIVITY, INSIDE_MAIN_ACTIVITY:
		{
			string_return += fmt.Sprintf("\nLANGS/EXTENTIONS: %s", LangInfo.languages)
			string_return += fmt.Sprintf("\n %s -> %s %s\n", m.spinner.View(), "Doing Language: ", LangInfo.languages[LangInfo.language_index])
		}
	}

	if m.state == ASKING_API_KEY || m.state == ASKING_LANGUAGE_TYPES {
		string_return += "\nPress ctrl+v to paste"
	}
	string_return += m.showFooter("\nPress ctrl+c to quit")
	return string_return
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Sorry, failed to start program, with error %s", err)
		os.Exit(1)
	}
}
