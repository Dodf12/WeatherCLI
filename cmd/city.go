/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cityCmd represents the city command
var cityCmd = &cobra.Command{
	Use:   "city",
	Short: "Type the city name to get the weather information for that day",
	Long: `WeatherCLI provides the weather information for the day in the city provided by the user.
Example:
	weatherCLI city --name "City Name"`,
	Run: func(cmd *cobra.Command, args []string) {
		cityName, _ := cmd.Flags().GetString("name")
		fmt.Println("Fetching your weather...", cityName)

		fmt.Println("---------------------------------------------------")

		lat, long := getLatLong(cityName)
		temperature, tempUnit, description := fetchWeatherData(lat, long)

		// Classify the weather and display ASCII art
		weatherKind := classifyWeather(description)
		displayArt(weatherKind.String(), temperature, tempUnit, description)

		fmt.Println("---------------------------------------------------")

	},
}

func init() {
	rootCmd.AddCommand(cityCmd)
	cityCmd.Flags().StringP("name", "n", "", "Name of the city to get weather information")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cityCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cityCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
