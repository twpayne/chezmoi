package cmd

import (
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"
)

type executeTemplateCmdConfig struct {
	init         bool
	promptString map[string]string
}

var executeTemplateCmd = &cobra.Command{
	Use:     "execute-template [templates...]",
	Short:   "Write the result of executing the given template(s) to stdout",
	Long:    mustGetLongHelp("execute-template"),
	Example: getExample("execute-template"),
	PreRunE: config.ensureNoError,
	RunE:    config.runExecuteTemplateCmd,
}

func init() {
	rootCmd.AddCommand(executeTemplateCmd)

	persistentFlags := executeTemplateCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.executeTemplate.init, "init", "i", false, "simulate chezmoi init")
	persistentFlags.StringToStringVarP(&config.executeTemplate.promptString, "promptString", "p", nil, "simulate promptString")
}

func (c *Config) runExecuteTemplateCmd(cmd *cobra.Command, args []string) error {
	if c.executeTemplate.init {
		c.templateFuncs["promptString"] = func(prompt string) string {
			if value, ok := c.executeTemplate.promptString[prompt]; ok {
				return value
			}
			return prompt
		}
	}

	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	switch len(args) {
	case 0:
		data, err := ioutil.ReadAll(c.Stdin)
		if err != nil {
			return err
		}
		result, err := ts.ExecuteTemplateData("stdin", data)
		if err != nil {
			return err
		}
		_, err = c.Stdout.Write(result)
		return err
	default:
		for i, arg := range args {
			result, err := ts.ExecuteTemplateData("arg"+strconv.Itoa(i+1), []byte(arg))
			if err != nil {
				return err
			}
			if _, err := c.Stdout.Write(result); err != nil {
				return err
			}
		}
		return nil
	}
}
