package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/zhangmingkai4315/go-storage/cmd"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "dataServer",
			Usage: "start a data server for object storage",
			Action: func(c *cli.Context) error {
				log.Println("starting data server...")
				cmd.RunDataServer()
				return nil
			},
		}, {
			Name:  "apiServer",
			Usage: "start a data server for object storage",
			Action: func(c *cli.Context) error {
				log.Println("starting api server...")
				cmd.RunAPIServer()
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
