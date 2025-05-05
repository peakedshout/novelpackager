package boot

import (
	"github.com/peakedshout/novelpackager/pkg/utils"
	"github.com/spf13/cobra"
)

const Version = `v0.1.2`

func Init(c *cobra.Command) {
	c.AddCommand(utils.ListCommand()...)
}
