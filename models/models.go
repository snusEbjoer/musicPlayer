package models

import "main/state"

type Model interface {
	WindowKey() state.ProgramWindow
}
