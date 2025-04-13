package utils

import "github.com/spf13/cobra"

var _cmdList []*cobra.Command

func RegisterCommand(c *cobra.Command) {
	_cmdList = append(_cmdList, c)
}

func ListCommand() []*cobra.Command {
	return _cmdList
}
