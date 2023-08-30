package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	ini "gopkg.in/ini.v1"
	lipgloss "github.com/charmbracelet/lipgloss"
	//wakatime_cli "github.com/wakatime/wakatime-cli/cmd"
)

type Status int
type Msg[T any] struct { value T}

const (
  STARTING Status = iota
  GOTTEN_HOME_DIR
  ASKING_API_KEY  
  GETTING_API_KEY
  GOTTEN_API_KEY
  ASKING_TYPES
)

type model struct {
  text_input textinput.Model
  wakatime_cfg string
  file_types []string
  file_type_index uint32
  state Status
}

func newModel() (model){
  text_in := textinput.New()
  return model{ state: STARTING, text_input: text_in }
}

func getHomeDir() tea.Msg{
  home, err := os.UserHomeDir()
  if err!=nil {
    fmt.Println("Couldn't get home dir of your os")
    return tea.Quit
  }

  var result Msg[string]
  result.value = home+string(os.PathSeparator)+".wakatime.test.cfg"
  return result
}

// Returns true if api-key field is set
// otherwise returns false
func (m model) checkIfApiKeySet() tea.Msg {
  cfg, err := ini.Load(m.wakatime_cfg)
  if err!=nil {
    //probably file doesn't exist so create it
    os.Create(m.wakatime_cfg)
    cfg_n, err := ini.Load(m.wakatime_cfg)
    if err!=nil {
      fmt.Printf("Error occured with loading config file: %s, with error: %s", m.wakatime_cfg, err)
      return tea.Quit
    }
    cfg = cfg_n
  }

  if !cfg.HasSection("settings") {
    cfg.NewSection("settings")
    cfg.SaveTo(m.wakatime_cfg)
  }
  settings := cfg.Section("settings")
  var result Msg[bool]
  if !settings.HasKey("api_key") {
    result.value=false
  }else{
    result.value=true
  }
  return result
}

func (m model) setApiKey(){
  cfg, err := ini.Load(m.wakatime_cfg)
  if err!=nil {
    fmt.Printf("Error, %s occured trying to open %s", err, m.wakatime_cfg)
  }

  cfg.Section("settings").NewKey("api_key", m.text_input.Value())
  if cfg.SaveTo(m.wakatime_cfg)!=nil {
    fmt.Printf("Failed to save %s file", m.wakatime_cfg)
  }
}

func (m model) Init() tea.Cmd {
  return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
    //handle key presses
    case tea.KeyMsg: {
      switch msg.Type {
        case tea.KeyCtrlC: return m, tea.Quit
	case tea.KeyEnter: {
	  if m.state == ASKING_API_KEY {
	    m.state = GOTTEN_API_KEY
	    m.setApiKey()
	    return m, textinput.Blink
	  }
	}
	case tea.KeyCtrlV: {
	  if m.state == ASKING_API_KEY {
	    return m, textinput.Paste
	  }
	}
      }
    }


    case Msg[string]: {
      switch m.state {
	case STARTING: {
	  m.state = GOTTEN_HOME_DIR
	  m.wakatime_cfg = msg.value
	  return m, m.checkIfApiKeySet
	}
      }
    }

    case Msg[bool]: {
      switch m.state {
	case GOTTEN_HOME_DIR: {
	  if msg.value == true {
	    m.state = GOTTEN_API_KEY
	    return m, nil
	  }else{
	    m.state = ASKING_API_KEY
	    m.text_input.Placeholder= "Enter Wakatime Api Key"
	    m.text_input.Width = 100
	    m.text_input.CharLimit = 100
	    m.text_input.Focus()
	    return m, textinput.Blink
	  }
	}
      }
    }
  }


  switch m.state{
    //do this when starting
    case STARTING: return m, getHomeDir
    case ASKING_API_KEY: {
      input, cmd := m.text_input.Update(msg);
      m.text_input = input
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
    Render("\nWelcome to Wakarizer\n")
}

func (m model) showTitle(title string) string {
  return lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#EEEEEE")).
    Render("\n"+title)
}

func (m model) showFooter(title string) string {
  return lipgloss.NewStyle().Bold(true).
    Foreground(lipgloss.Color("#EE1111")).
    Render(title)
}

func (m model) View() string {
  string_return := m.showMainTitle()

  switch m.state {
    case STARTING: string_return += m.showTitle("\nLoading....")
    case GOTTEN_HOME_DIR: string_return += m.showTitle(fmt.Sprintf("\nGotten config @ %s", m.wakatime_cfg))
    case ASKING_API_KEY: {
      string_return += fmt.Sprintf("%s\n\n%s",m.showTitle("\nWhat is your api key?"), m.text_input.View())
    }
    case GOTTEN_API_KEY: string_return += m.showTitle("\nSet api key")
  }

  if m.state == ASKING_API_KEY {
    string_return += "\nPress ctrl+v to paste"
  }
  string_return += m.showFooter("\nPress ctrl+c to quit ")
  return string_return
}

func main(){
  p := tea.NewProgram(newModel())
  if _, err := p.Run(); err != nil {
    fmt.Printf("Sorry, failed to start program, with error %s",err)
    os.Exit(1)
  }
}
