/*
Copyright © 2019 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"doppler-cli/api"
	configuration "doppler-cli/config"
	"doppler-cli/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type secretsResponse struct {
	Variables map[string]interface{}
	Success   bool
}

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Fetch all Doppler secrets",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Println("Error: no argument expected")
			cmd.Help()
			return
		}

		plain, err := strconv.ParseBool(cmd.Flag("plain").Value.String())
		if err != nil {
			utils.Err(err)
		}

		jsonFlag, err := strconv.ParseBool(cmd.Flag("json").Value.String())
		if err != nil {
			utils.Err(err)
		}

		raw, err := strconv.ParseBool(cmd.Flag("raw").Value.String())
		if err != nil {
			utils.Err(err)
		}

		localConfig := configuration.LocalConfig(cmd)

		if jsonFlag {
			response, _ := api.GetAPISecrets(cmd, localConfig.Key.Value, localConfig.Project.Value, localConfig.Config.Value, false)
			fmt.Println(string(response))
			return
		}

		_, secrets := api.GetAPISecrets(cmd, localConfig.Key, localConfig.Project, localConfig.Config, true)

		var matchedSecrets []api.ComputedSecret
		for _, name := range args {
			if secrets[name] != (api.ComputedSecret{}) {
				matchedSecrets = append(matchedSecrets, secrets[name])
			}
		}

		if plain {
			sbEmpty := true
			var sb strings.Builder
			for _, secret := range matchedSecrets {
				if sbEmpty {
					sbEmpty = false
				} else {
					sb.WriteString("\n")
				}

				if raw {
					sb.WriteString(secret.RawValue)
				} else {
					sb.WriteString(secret.ComputedValue)
				}
			}

			fmt.Println(sb.String())
			return
		}

		// TODO doesn't handle multi line secrets well at all
		table := tablewriter.NewWriter(os.Stdout)
		headers := []string{"name", "value"}
		if raw {
			headers = append(headers, "raw")
		}
		table.SetHeader(headers)

		for _, secret := range secrets {
			row := []string{secret.Name, secret.ComputedValue}
			if raw {
				row = append(row, secret.RawValue)
			}
			table.Append(row)
		}

		table.Render()
	},
}

var secretsGetCmd = &cobra.Command{
	Use:   "get [secrets]",
	Short: "Get the value of one or more secrets",
	Long: `Get the value of one or more secrets.

Ex: output the secrets "api_key" and "crypto_key":
doppler secrets get api_key crypto_key`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: missing argument")
			cmd.Help()
			return
		}

		plain, err := strconv.ParseBool(cmd.Flag("plain").Value.String())
		if err != nil {
			utils.Err(err)
		}

		raw, err := strconv.ParseBool(cmd.Flag("raw").Value.String())
		if err != nil {
			utils.Err(err)
		}

		jsonFlag, err := strconv.ParseBool(cmd.Flag("json").Value.String())
		if err != nil {
			utils.Err(err)
		}

		localConfig := configuration.LocalConfig(cmd)
		_, secrets := api.GetAPISecrets(cmd, localConfig.Key.Value, localConfig.Project.Value, localConfig.Config.Value, true)

		if jsonFlag {
			secretsMap := make(map[string]map[string]string)
			for _, name := range args {
				if secrets[name] != (api.ComputedSecret{}) {
					secretsMap[name] = make(map[string]string)
					secretsMap[name]["raw"] = secrets[name].RawValue
					secretsMap[name]["computed"] = secrets[name].ComputedValue
				}
			}

			resp, err := json.Marshal(secretsMap)
			if err != nil {
				utils.Err(err)
			}

			fmt.Println(string(resp))
			return
		}

		var matchedSecrets []api.ComputedSecret
		for _, name := range args {
			if secrets[name] != (api.ComputedSecret{}) {
				matchedSecrets = append(matchedSecrets, secrets[name])
			}
		}

		if plain {
			sbEmpty := true
			var sb strings.Builder
			for _, secret := range matchedSecrets {
				if sbEmpty {
					sbEmpty = false
				} else {
					sb.WriteString("\n")
				}

				if raw {
					sb.WriteString(secret.RawValue)
				} else {
					sb.WriteString(secret.ComputedValue)
				}
			}

			fmt.Println(sb.String())
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		headers := []string{"name", "value"}
		if raw {
			headers = append(headers, "raw")
		}
		table.SetHeader(headers)

		for _, secret := range matchedSecrets {
			row := []string{secret.Name, secret.ComputedValue}
			if raw {
				row = append(row, secret.RawValue)
			}
			table.Append(row)
		}

		table.Render()
	},
}

var secretsDownloadCmd = &cobra.Command{
	Use:   "download <filename>",
	Short: "Download a config's .env file",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, err := filepath.Abs(cmd.Flag("path").Value.String())
		if err != nil {
			utils.Err(err)
		}
		fileName := "doppler.env"
		if len(args) > 0 {
			fileName = args[0]
		}

		var metadata bool
		metadata, err = strconv.ParseBool(cmd.Flag("metadata").Value.String())
		if err != nil {
			metadata = true
		}

		localConfig := configuration.LocalConfig(cmd)
		body := api.DownloadSecrets(cmd, localConfig.Key.Value, localConfig.Project.Value, localConfig.Config.Value, metadata)

		err = ioutil.WriteFile(path.Join(filePath, fileName), body, 0600)
		if err != nil {
			utils.Err(err)
		}
	},
}

func init() {
	secretsCmd.Flags().Bool("raw", false, "display the raw secret value without processing variables")
	secretsCmd.Flags().Bool("plain", false, "print values without formatting")

	secretsGetCmd.Flags().Bool("plain", false, "print values without formatting")
	secretsGetCmd.Flags().Bool("raw", false, "display the raw secret value without processing variables")

	secretsDownloadCmd.Flags().String("path", ".", "location to save the file")
	secretsDownloadCmd.Flags().Bool("metadata", true, "add metadata to the downloaded file (helps cache busting)")

	secretsCmd.AddCommand(secretsGetCmd)
	secretsCmd.AddCommand(secretsDownloadCmd)
	rootCmd.AddCommand(secretsCmd)
}