package cmd

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "17clock",
	Short: "clock in & off",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			color.Red(err.Error())
			return
		}

		employee := viper.GetString("employee")
		identity := viper.GetString("identity")
		latitude := viper.GetString("latitude")
		longitude := viper.GetString("longitude")
		address := viper.GetString("address")

		sessionID := login(employee, identity)

		getTodayRecord(sessionID)

		list()

		var id string
		prompt := &survey.Input{
			Message: "Enter Action ID",
			Default: "",
		}

		if err := survey.AskOne(prompt, &id); err != nil {
			color.Red(err.Error())
			return
		}

		switch id {
		case "1":
			clockIn(sessionID, latitude, longitude, address)
		case "2":
			clockOff(sessionID, latitude, longitude, address)
		default:
			color.Red("Invalid Action")
		}
	},
}

func list() {
	rows := [][]string{
		{"1", "上班打卡"},
		{"2", "下班打卡"},
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Action ID", "Action"})
	table.AppendBulk(rows)
	table.SetColWidth(1000)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	table.Render()
}

var cmdAuto = &cobra.Command{
	Use:   "auto",
	Short: "auto clock in & off",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			color.Red(err.Error())
			return
		}

		if isTodayHoliday() {
			color.Red("Today is holiday!")
			return
		}

		whichTime := getClockInOrClockOff()
		employee := viper.GetString("employee")
		identity := viper.GetString("identity")
		latitude := viper.GetString("latitude")
		longitude := viper.GetString("longitude")
		address := viper.GetString("address")

		sessionID := login(employee, identity)

		getTodayRecord(sessionID)

		switch whichTime {
		case clockInTime:
			clockIn(sessionID, latitude, longitude, address)
		case clockOffTime:
			clockOff(sessionID, latitude, longitude, address)
		default:
			color.Red("It's not clockIn or clockOff time")
		}
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	rootCmd.AddCommand(cmdAuto)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		return
	}
}
