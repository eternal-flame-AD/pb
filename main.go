package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/eternal-flame-AD/go-pushbullet"
	"github.com/olekukonko/tablewriter"
)

func ErrAndExit(err interface{}, code int) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}

func init() {
	if ExistConfig() {
		if err := ReadConfig(); err != nil {
			panic(err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Config file does not exist... Creating")
		config = Config{}
		if err := WriteConfig(); err != nil {
			panic(err)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		ErrAndExit("No subcommand specified!\n Available subcommands: \n\tconfig\n\tdevice\n\tpush", 2)
	}
	subcommand := os.Args[1]
	switch subcommand {
	case "config":
		keys := make([]string, 0)
		ConfigType := reflect.TypeOf(config)
		for i := 0; i < ConfigType.NumField(); i++ {
			keys = append(keys, ConfigType.Field(i).Tag.Get("key"))
		}
		if len(os.Args) < 3 {
			ErrAndExit("No subcommand specified!", 2)
		}
		subcommand := os.Args[2]
		switch subcommand {
		case "show":
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"name", "value", "explanation"})
			for index, key := range keys {
				help, _ := reflect.TypeOf(config).Field(index).Tag.Lookup("help")
				table.Append([]string{key, reflect.ValueOf(config).Field(index).String(), help})
			}
			table.Render()
		case "set":
			if len(os.Args) != 5 {
				ErrAndExit("Usage: pb config set <name> <value>", 2)
			}
			key, value := os.Args[3], os.Args[4]
			ok := false
			for index, thiskey := range keys {
				if key == thiskey {
					reflect.ValueOf(&config).Elem().Field(index).SetString(value)
					ok = true
				}
			}
			if ok {
				if err := WriteConfig(); err != nil {
					ErrAndExit("Error while writing config file: "+err.Error(), 3)
				}
				fmt.Println("Successfully updated config file!")
			} else {
				ErrAndExit("Config entry not found", 2)
			}
		}
	case "device":
		if len(os.Args) < 3 {
			ErrAndExit("No subcommand specified!\navailable subcommands:\n\tlist", 2)
		}
		client := pushbullet.New(config.Key)
		switch os.Args[2] {
		case "list":
			flagset := flag.NewFlagSet("device list", flag.ExitOnError)
			verbosevar := boolflag{new(bool)}
			flagset.Var(verbosevar, "v", "verbose")
			flagset.Parse(os.Args[3:])
			verbose := verbosevar.Get().(bool)
			devices, err := client.Devices()
			if err != nil {
				ErrAndExit(err, 5)
			}
			if len(devices) == 0 {
				ErrAndExit("No devices available!", 5)
			}
			skipfields := map[string]bool{"Client": true, "Fingerprint": false, "KeyFingerprint": false, "PushToken": false}
			fields := make([]string, 0)
			index := 0
			for i := 0; i < reflect.TypeOf(*devices[1]).NumField(); i++ {
				name := reflect.TypeOf(*devices[1]).Field(i).Name
				if skipInVerbose, ok := skipfields[name]; ok && (skipInVerbose || !verbose) {
					continue
				}
				fields = append(fields, name)
				index++
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader(append([]string{"index"}, fields...))
			for index, device := range devices {
				values := []string{strconv.Itoa(index)}
				for _, name := range fields {
					if skipInVerbose, ok := skipfields[name]; ok && (skipInVerbose || !verbose) {
						continue
					}
					values = append(values, fmt.Sprint(reflect.ValueOf(device).Elem().FieldByName(name).Interface()))
				}
				table.Append(values)
			}
			table.Render()
		default:
			ErrAndExit(fmt.Sprintf("Subcommand %s does not exist", os.Args[2]), 2)
		}
	case "push":
		if len(os.Args) < 3 {
			ErrAndExit("No type specified!Available types:\n\tnote\n\tlink", 2)
		}
		pushtype := os.Args[2]
		client := pushbullet.New(config.Key)
		switch pushtype {
		case "note":
			flagset := flag.NewFlagSet("pushnote", flag.ExitOnError)
			device := flagset.String("d", "", "target device. Parse order: index, Nickname, Model, Iden")
			title := flagset.String("t", "", "message title")
			message := flagset.String("m", "", "note message")
			flagset.Parse(os.Args[3:])
			devicekey := func() string {
				if *device == "" {
					return ""
				}
				res, err := locateDeviceWithClient(*device, client)
				if err != nil {
					ErrAndExit(err, 2)
				}
				return res
			}()
			if err := client.PushNote(devicekey, *title, *message); err != nil {
				ErrAndExit(err, 5)
			}
			fmt.Println("Success!")
		case "link":
			flagset := flag.NewFlagSet("pushlink", flag.ExitOnError)
			device := flagset.String("d", "", "target device. Parse order: index, Nickname, Model, Iden")
			title := flagset.String("u", "", "link URL")
			message := flagset.String("m", "", "message")
			flagset.Parse(os.Args[3:])
			devicekey := func() string {
				if *device == "" {
					return ""
				}
				res, err := locateDeviceWithClient(*device, client)
				if err != nil {
					ErrAndExit(err, 2)
				}
				return res
			}()
			if err := client.PushNote(devicekey, *title, *message); err != nil {
				ErrAndExit(err, 5)
			}
			fmt.Println("Success!")
		default:
			ErrAndExit("Push type unsupported", 2)
		}
	case "listen":
		client := pushbullet.New(config.Key)
		listener := client.Listen()
		for {
			select {
			case push := <-listener.Push:
				log.Printf("New %s from %s: %s\n%s\n", push.Type, push.SenderName, push.Title, push.Body)
				switch push.Type {
				case "note":
				case "file":
					fmt.Printf("File %s of type %s: %s\n", push.FileName, push.FileMIME, push.FileURL)
				case "link":
					fmt.Printf("Link: %s\n", push.URL)
				}
			case ephemeral := <-listener.Ephemeral:
				log.Printf("New ephemeral of type %s received from %s@%s: %s\n", ephemeral.Type, ephemeral.PackageName, ephemeral.SourceUserIden, ephemeral.Message)
			case device := <-listener.Device:
				log.Printf("Device %s updated\n", device.Nickname)
			case err := <-listener.Error:
				log.Printf("An error occured: %s", err.Error())
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Invalid subcommand: %s\n", subcommand)
		os.Exit(2)
	}

}
