package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ini "gopkg.in/ini.v1"
	"wakarizer/wakatime"
)

type Status int
type Msg[T any] struct { value T}

const (
  STARTING Status = iota
  GOTTEN_HOME_DIR
  ASKING_API_KEY  
  SETUP_ASKING_LANGUAGE_TYPES
  ASKING_LANGUAGE_TYPES
  VALIDATE_GETTING_LANGUAGES
  START_MAIN_ACTIVITY
)

type model struct {
  text_input textinput.Model
  wakatime_cfg string
  file_types []string
  file_type_index uint32
  state Status
  languages []string
  language_index int
  spinner spinner.Model
}

func newModel() (model){
  spin := spinner.New()
  spin.Spinner = spinner.Dot
  spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#11EE11"))

  return model{ state: STARTING,
    text_input: textinput.New(),
    spinner: spin,
    language_index: 0,
  }
}

func getHomeDir() tea.Msg{
  home, err := os.UserHomeDir()
  if err!=nil {
    fmt.Println("Couldn't get home dir of your os")
    return tea.Quit
  }

  var result Msg[string]
  result.value = home+string(os.PathSeparator)+".wakarizer.cfg"
  return result
}

func (m model) getKey() string {
  cfg,  err := ini.Load(m.wakatime_cfg)
  if err!=nil {
    fmt.Printf("Couldn't load cfg file from %s\n", m.wakatime_cfg)
    os.Exit(1)
  }

  return cfg.Section("settings").Key("api_key").String()
}

// Perform the actual wakatime hearbeat
func (m *model) doHeartBeat() {
  key := m.getKey()
  wakatime.Execute(m.languages[m.language_index], key)
  m.updateLanguageIndex()
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

func (m *model) updateLanguageIndex(){
  new_val := m.language_index+1
  if( new_val == len(m.languages)){
    m.language_index = 0
  }else{
    m.language_index = new_val
  }
}

func (m model) checkIfLangsInputEmpty() tea.Msg {
  var result Msg[bool]
  for _, v := range m.languages {
    if len(v) == 0 {
      result.value=false
      return result
    }
  }
  result.value=true
  return result
}

func (m *model) setApiKey(){
  cfg, err := ini.Load(m.wakatime_cfg)
  if err!=nil {
    fmt.Printf("Error, %s occured trying to open %s", err, m.wakatime_cfg)
    os.Exit(1)
  }

  cfg.Section("settings").NewKey("api_key", m.text_input.Value())
  if cfg.SaveTo(m.wakatime_cfg)!=nil {
    fmt.Printf("Failed to save %s file", m.wakatime_cfg)
    os.Exit(1)
  }
}

func (m *model) setLanguages(){
  m.languages = strings.Split(m.text_input.Value(), " ")
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
	  switch m.state {
	    case ASKING_API_KEY: {
	      m.state = SETUP_ASKING_LANGUAGE_TYPES
	      m.setApiKey()
	      return m, textinput.Blink
	    }
	    case ASKING_LANGUAGE_TYPES: {
	      m.state = VALIDATE_GETTING_LANGUAGES
	      m.setLanguages()
	      return m, m.checkIfLangsInputEmpty
	    }
	  }
	}
	case tea.KeyCtrlV: {
	  if m.state == ASKING_API_KEY || m.state == ASKING_LANGUAGE_TYPES {
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
	    m.state = SETUP_ASKING_LANGUAGE_TYPES
	    return m, textinput.Blink
	  }else{
	    m.state = ASKING_API_KEY
	    m.text_input.Placeholder= "Enter Wakatime Api Key"
	    m.text_input.Width = 100
	    m.text_input.CharLimit = 100
	    m.text_input.Focus()
	    return m, textinput.Blink
	  }
	}
	case VALIDATE_GETTING_LANGUAGES: {
	  if msg.value==true {
	    m.state = START_MAIN_ACTIVITY
	    return m, m.spinner.Tick
	  }else{
	    m.state = ASKING_LANGUAGE_TYPES
	    return m, textinput.Blink
	  }
	}
      }
    }
  }


  switch m.state{
    //do this when starting
    case STARTING: return m, getHomeDir
    case ASKING_API_KEY, ASKING_LANGUAGE_TYPES: {
      input, cmd := m.text_input.Update(msg);
      m.text_input = input
      return m, cmd
    }
    case SETUP_ASKING_LANGUAGE_TYPES: {
      m.state = ASKING_LANGUAGE_TYPES
      m.text_input = textinput.New()
      m.text_input.Placeholder= "rs ts java...example of how to enter file extentions"
      m.text_input.Width = 100
      m.text_input.Focus()
      return m, textinput.Blink
    }
    case START_MAIN_ACTIVITY: {
      m.doHeartBeat()
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
    case ASKING_LANGUAGE_TYPES: string_return += fmt.Sprintf("%s\n\n%s",
      m.showTitle("\nWhat Languages You want to wakarize, give atleast 1? \n"+
      "(write the extention, and seperate with spaces for others\n"+
      "eg: rs ts java\n"+
      "This is for rust, typescript and java)"), m.text_input.View())

    case START_MAIN_ACTIVITY: {
      string_return += fmt.Sprintf("\nLANGS/EXTENTIONS: %s", m.languages)
      string_return += fmt.Sprintf("\n %s -> %s %s\n", m.spinner.View(), "Doing Language: ", m.languages[m.language_index])
    }
  }

  if m.state == ASKING_API_KEY || m.state == ASKING_LANGUAGE_TYPES {
    string_return += "\nPress ctrl+v to paste"
  }
  if m.state == START_MAIN_ACTIVITY {
  string_return +=  "\nBOOM:=>"+m.getKey()
  }
  string_return += m.showFooter("\nPress ctrl+c to quit")
  return string_return
}

func main(){
  p := tea.NewProgram(newModel())
  if _, err := p.Run(); err != nil {
    fmt.Printf("Sorry, failed to start program, with error %s",err)
    os.Exit(1)
  }
}
