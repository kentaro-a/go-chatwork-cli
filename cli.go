package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kr/pretty"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
)

var app *cli.App

type Config struct {
	Name     string `json:"name"`
	ApiToken string `json:"apitoken"`
}

func main() {
	app = cli.NewApp()
	app.Name = "Simple Chatwork Sender"
	app.Usage = "cw options message file"
	app.Version = "1.0.0"
	app.Commands = []cli.Command{
		{
			Name:    "configure",
			Aliases: []string{"c"},
			Usage:   "Configure someting to need to exec",
			Action:  configure,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "list, l",
				},
			},
		},
		{
			Name:   "rooms",
			Usage:  "Show rooms and their id",
			Action: rooms,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "apitoken, at",
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "apitoken",
		},
		cli.StringFlag{
			Name: "to",
		},
		cli.StringFlag{
			Name: "message, m",
		},
		cli.StringFlag{
			Name: "tpl",
		},
		cli.StringFlag{
			Name: "file, f",
		},
	}
	app.Action = send
	app.HideHelp = true
	app.Run(os.Args)
}

func send(c *cli.Context) error {
	config := Config{}
	if apitoken := c.String("apitoken"); apitoken != "" {
		config.ApiToken = apitoken
	} else {
		if fpath := getConfigFile(&config); fpath == "" {
			fmt.Println("Require to set apitoken by at least --apitoken or config file")
			return nil
		}
	}
	cw := Api{ApiToken: config.ApiToken}

	to := c.String("to")
	if to == "" {
		fmt.Println("Please set sendto roomname with --to")
		return nil
	}

	msg := ""
	_tpl := c.String("tpl")
	if b, err := ioutil.ReadFile(_tpl); err == nil {
		msg = string(b)
	} else {
		msg = c.String("message")
	}
	if msg == "" {
		fmt.Println("Please set message or message template with --message, --tpl")
		return nil
	}

	attach_file_path := c.String("file")
	if attach_file_path != "" {
		if f, err := os.Stat(attach_file_path); os.IsNotExist(err) {
			fmt.Println("File not found: ", attach_file_path)
			return nil
		} else if float64(f.Size()) > float64(5*1024*1024) {
			fmt.Println("FileSize is accepted up to 5M")
			return nil
		}

	}

	var res []byte
	var err error
	if attach_file_path != "" {
		res, err = cw.SendMessageByNameWithFile(to, msg, attach_file_path)
	} else {
		res, err = cw.SendMessageByName(to, msg)
	}
	if err != nil {
		fmt.Println("------ Error -------")
		fmt.Println(err)
		fmt.Println("------ Response -------")
		fmt.Println(string(res))
	} else {
		fmt.Println("Successfully Sent.")
	}
	return nil
}

func rooms(c *cli.Context) error {
	config := Config{}
	if apitoken := c.String("apitoken"); apitoken != "" {
		config.ApiToken = apitoken
	} else {
		if fpath := getConfigFile(&config); fpath == "" {
			fmt.Println("Require to set apitoken by at least --apitoken or config file")
			return nil
		}
	}
	cw := Api{ApiToken: config.ApiToken}
	if err := cw.GetRooms(); err != nil {
		fmt.Println("Cannot get rooms")
		return nil
	}
	for k, v := range cw.RoomHash {
		fmt.Printf("\nRoomID: %d, Name: %s", v, k)
	}
	fmt.Println("")
	return nil
}

func configure(c *cli.Context) error {
	if c.Bool("list") {
		config := Config{}
		fpath := getConfigFile(&config)
		if fpath == "" {
			fmt.Println("Not configured yet. Please configure before using.")
			return nil
		} else {
			fmt.Printf("\n%# v\n", pretty.Formatter(config))
			return nil
		}

	} else {
		s := bufio.NewScanner(os.Stdin)
		name := ""
		apitoken := ""
		for {
			fmt.Printf("Put your name: ")
			s.Scan()
			name = s.Text()
			if name != "" {
				break
			}
		}
		for {
			fmt.Printf("Put your apitoken: ")
			s.Scan()
			apitoken = s.Text()
			if apitoken != "" {
				break
			}
		}
		config := Config{
			Name:     name,
			ApiToken: apitoken,
		}
		json, err := json.MarshalIndent(config, "", "\t")
		if err != nil {
			fmt.Println("Cannot encode to configfile")
			return nil
		}
		fpath := filepath.Join(".", ".config", "config.json")
		fp, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			fmt.Println("Cannot save config as file: ", fpath)
			return nil
		}
		defer fp.Close()
		if _, err := fp.Write(json); err != nil {
			fmt.Println("Cannot save config as file: ", fpath)
			return nil
		}
	}
	return nil
}

func getConfigFile(config *Config) (fpath string) {
	fpath = filepath.Join(".", ".config", "config.json")
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		fpath = ""
	} else {
		if err := json.Unmarshal(b, config); err != nil {
			fpath = ""
		}
	}
	return
}
