// Copyright © 2017 Charles Burton charles.d.burton@gmail.com
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/TheJumpCloud/jcapi"
	"github.com/spf13/cobra"
)

var (
	serverName   = ""
	sharedSecret = ""
	delete       = false
)

// radiusCmd represents the radius command
var radiusCmd = &cobra.Command{
	Use:   "radius",
	Short: "Create Update or Delete Radius Server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if apiKey == "" || serverName == "" {
			log.Println("API Key or Name not set!")
			os.Exit(0)
		}

		jcAPI := jcapi.NewJCAPI(apiKey, jcapi.StdUrlBase)
		if !delete {
			err := createOrUpdateServer(jcAPI)
			if err != nil {
				panic(err)
			}
		} else {
			err := deleteServer(jcAPI)
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(radiusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// radiusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// radiusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	radiusCmd.Flags().StringVarP(&serverName, "name", "n", "", "Set the name")
	radiusCmd.Flags().StringVarP(&sharedSecret, "secret", "s", "", "The shared radius secret")
	radiusCmd.Flags().BoolVarP(&delete, "delete", "d", false, "Name of server to delete")
}

func createOrUpdateServer(api jcapi.JCAPI) error {
	server, err := getServerByName(api, serverName)
	if err != nil {
		return err
	}
	if server.Name == serverName {
		err := updateServer(server, api)
		if err != nil {
			return err
		}
	} else {
		err := createServer(api)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteServer(api jcapi.JCAPI) error {
	server, err := getServerByName(api, serverName)
	if err != nil {
		return err
	}
	if server.Name == serverName {
		err := api.DeleteRadiusServer(server)
		if err != nil {
			return err
		}
	} else {
		log.Println("Server: ", serverName, " not found")
	}
	return nil
}

//Retrieve your externally facing IP Address
func getIp() (string, error) {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		log.Println(err)
		return "", err

	}
	if resp.StatusCode == 200 {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			return "", err2
		}
		bodyString := string(bodyBytes)
		return strings.TrimSpace(bodyString), nil
	}
	return "", errors.New("Unable to retrieve external Ip")
}

//Search for the Radius server configured with the given name
func getServerByName(jc jcapi.JCAPI, name string) (jcapi.JCRadiusServer, error) {
	var dummy jcapi.JCRadiusServer
	jcServers, err := jc.GetAllRadiusServers()
	if err != nil {
		return dummy, err
	}
	for _, server := range jcServers {
		if server.Name == serverName {
			return server, nil
		}
	}
	return dummy, nil
}

//If a server was found, update its current state(mostly the ip address)
func updateServer(server jcapi.JCRadiusServer, api jcapi.JCAPI) error {
	ip, err := getIp()
	if err != nil {
		return err
	}
	server.NetworkSourceIP = strings.TrimSpace(ip)
	id, err := api.AddUpdateRadiusServer(jcapi.Update, server)
	if err != nil {
		return err
	}
	log.Println(id)
	return nil
}

//If a server was not found with that name, create a new one
func createServer(api jcapi.JCAPI) error {
	log.Println("Name not found, creating new")
	var server jcapi.JCRadiusServer
	if sharedSecret == "" {
		return errors.New("Shared Secret not set")
	}
	ip, err := getIp()
	log.Println("Got IP: ", ip)
	if err != nil {
		return err
	}
	server.NetworkSourceIP = strings.TrimSpace(ip) //Remove whitspace, this sucked
	server.Name = serverName
	server.SharedSecret = sharedSecret
	id, err := api.AddUpdateRadiusServer(jcapi.Insert, server)
	if err != nil {
		return err
	}
	log.Println(id)
	return nil
}
