package main

import (
	"code.google.com/p/winsvc/eventlog"
	"code.google.com/p/winsvc/mgr"
	"code.google.com/p/winsvc/svc"
	"fmt"
	"mongodb.com/munin-agent/components"
	"os"
	"path/filepath"
)

const MUNINNAME = "MMSMunin"
const MUNINDISPLAYNAME = "MMS Munin Agent"
const MUNINDESCRIPTION = "Munin agent for MongoDB Management Service."

type muninService struct {}

func (self *muninService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	eventlog.InstallAsEventCreate(MUNINDISPLAYNAME, eventlog.Info|eventlog.Warning|eventlog.Error)
	elog, err := eventlog.Open(MUNINDISPLAYNAME)
	if err != nil {
		return
	}
	elog.Info(1, "Starting Munin agent")

	agent := &components.Agent{}
	go agent.Run()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop}
	for {
		c := <-r
		switch c.Cmd {
		case svc.Stop:
			changes <- svc.Status{State: svc.StopPending}

			return
		}
	}
}

func main() {
	exePath, _ := filepath.Abs(os.Args[0])
	if filepath.Ext(exePath) == "" {
		exePath += ".exe"
	}
	interactive, _ := svc.IsAnInteractiveSession()
	if !interactive {
		svc.Run(MUNINNAME, &muninService{})
		return
	}
	if len(os.Args) >= 2 && os.Args[1] == "install" {
		m, err := mgr.Connect()
		if err != nil {
			fmt.Printf("Error: Failed to connect to service control manager!\n")
			return
		}
		defer m.Disconnect()
		s, err := m.OpenService(MUNINNAME)
		if err == nil {
			s.Close()
			fmt.Printf("Service already installed!\n")
			return
		}
		s, err = m.CreateService(MUNINNAME, exePath, mgr.Config{StartType: mgr.StartAutomatic, DisplayName: MUNINDISPLAYNAME, Description: MUNINDESCRIPTION})
		if err != nil {
			fmt.Printf("Error: Failed to create service!\n")
			return
		}
		defer s.Close()
		return
	}
}
