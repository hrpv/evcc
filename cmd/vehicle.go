package cmd

import (
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/server"
	"github.com/evcc-io/evcc/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// vehicleCmd represents the vehicle command
var vehicleCmd = &cobra.Command{
	Use:   "vehicle [name]",
	Short: "Query configured vehicles",
	Run:   runVehicle,
}

func init() {
	rootCmd.AddCommand(vehicleCmd)
	vehicleCmd.PersistentFlags().StringP(flagName, "n", "", "select vehicle by name")
	vehicleCmd.PersistentFlags().BoolP(flagStart, "a", false, "start charge")
	vehicleCmd.PersistentFlags().BoolP(flagStop, "o", false, "stop charge")
	vehicleCmd.PersistentFlags().BoolP(flagWakeup, "w", false, flagWakeup)
}

func runVehicle(cmd *cobra.Command, args []string) {
	util.LogLevel(viper.GetString("log"), viper.GetStringMapString("levels"))
	log.INFO.Printf("evcc %s", server.FormattedVersion())

	// load config
	conf, err := loadConfigFile(cfgFile)
	if err != nil {
		log.FATAL.Fatal(err)
	}

	// setup environment
	if err := configureEnvironment(conf); err != nil {
		log.FATAL.Fatal(err)
	}

	// select single charger
	if name := cmd.PersistentFlags().Lookup(flagName).Value.String(); name != "" {
		for _, cfg := range conf.Vehicles {
			if cfg.Name == name {
				conf.Vehicles = []qualifiedConfig{cfg}
				break
			}
		}
	}

	if err := cp.configureVehicles(conf); err != nil {
		log.FATAL.Fatal(err)
	}

	vehicles := cp.vehicles
	if len(args) == 1 {
		arg := args[0]
		vehicles = map[string]api.Vehicle{arg: cp.Vehicle(arg)}
	}

	d := dumper{len: len(vehicles)}

	var flagUsed bool
	for _, v := range vehicles {
		if cmd.PersistentFlags().Lookup(flagWakeup).Changed {
			flagUsed = true

			if vv, ok := v.(api.AlarmClock); ok {
				if err := vv.WakeUp(); err != nil {
					log.ERROR.Println("wakeup:", err)
				}
			} else {
				log.ERROR.Println("wakeup: not implemented")
			}
		}

		if cmd.PersistentFlags().Lookup(flagStart).Changed {
			flagUsed = true

			if vv, ok := v.(api.VehicleStartCharge); ok {
				if err := vv.StartCharge(); err != nil {
					log.ERROR.Println("start charge:", err)
				}
			} else {
				log.ERROR.Println("start charge: not implemented")
			}
		}

		if cmd.PersistentFlags().Lookup(flagStop).Changed {
			flagUsed = true

			if vv, ok := v.(api.VehicleStopCharge); ok {
				if err := vv.StopCharge(); err != nil {
					log.ERROR.Println("stop charge:", err)
				}
			} else {
				log.ERROR.Println("stop charge: not implemented")
			}
		}
	}

	if !flagUsed {
		for name, v := range vehicles {
			d.DumpWithHeader(name, v)
		}
	}
}
