package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	//ini "gopkg.in/ini.v1"
	//lipgloss "github.com/charmbracelet/lipgloss"
	//wakatime_cli "github.com/wakatime/wakatime-cli/cmd"
)

type Status int
type MsgString[T any] struct { value T}

const (
  STARTING Status = iota
  GOTTEN_HOME_DIR
  TRY_GET_API_KEY
  ASKING_API_KEY  
  ASKING_TYPES
)

type model struct {
  wakatime_cfg string
  file_types []string
  file_type_index uint32
  state Status
}

func newModel() (model){
  return model{ state: STARTING }
}

func getHomeDir() tea.Msg{
  home, err := os.UserHomeDir()
  if err!=nil {
    fmt.Println("Couldn't get home dir of your os")
    return tea.Quit
  }

  var result MsgString[string]
  result.value = home+string(os.PathSeparator)+".wakatime.cfg"
  return result
}


//func (m *model) startSetup() {
//
//  cfg, err := ini.Load(m.wakatime_cfg)
//  if err!=nil {
//    //probably file doesn't exist so create it
//    os.Create(m.wakatime_cfg)
//    cfg_n, err := ini.Load(m.wakatime_cfg)
//    if err!=nil {
//      fmt.Printf("Error occured with loading config file: %s, with error: %s", m.wakatime_cfg, err)
//      return tea.Quit
//    }
//    cfg = cfg_n
//  }
//
//  if !cfg.HasSection("settings") {
//    cfg.NewSection("settings")
//    cfg.SaveTo(m.wakatime_cfg)
//  }
//  settings := cfg.Section("settings")
//  if !settings.HasKey("api-key") {
//    m.state = ASKING_API_KEY
//  }
//}

func (m model) Init() tea.Cmd {
  return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
    //handle key presses
    case tea.KeyMsg: {
      switch msg.String() {
        case "ctrl+c", "q":
          return m, tea.Quit
      }
    }

    case MsgString[string]: {
      switch m.state {
	case STARTING: {
	  m.state = GOTTEN_HOME_DIR
	  m.wakatime_cfg = msg.value
	  return m, nil
	}
      }
    }

    default : {
      //do this when starting
      if m.state == STARTING {
	return m, getHomeDir
      }
    }
  }
  return m, nil
}

func (m model) View() string {
  string_return := ""

  switch m.state {
    case STARTING: string_return += "Loading...."
    case GOTTEN_HOME_DIR: string_return += fmt.Sprintf("Gotten config @ %s", m.wakatime_cfg)
    case ASKING_API_KEY: string_return = "Asking for API_KEY"
  }

  string_return += "\nPress q to quit "
  return string_return
}

func main(){
  p := tea.NewProgram(newModel(), tea.WithAltScreen())
  if _, err := p.Run(); err != nil {
    fmt.Printf("Sorry, failed to start program, with error %s",err)
    os.Exit(1)
  }
}
