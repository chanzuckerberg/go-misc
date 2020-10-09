package cmds

import (
	"os"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	quiet      bool
	cpuprofile string
)

func Root(name, short, long string) *cobra.Command {
	root := &cobra.Command{
		Use:   name,
		Short: short,
		Long:  long,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel := logrus.InfoLevel
			if debug { // debug overrides quiet
				logLevel = logrus.DebugLevel
			} else if quiet {
				logLevel = logrus.FatalLevel
			}
			logrus.SetLevel(logLevel)
			logrus.Infof("log level set to %s", logrus.GetLevel())

			if cpuprofile != "" {
				logrus.Info("starting cpu profile")
				f, err := os.Create(cpuprofile)
				if err != nil {
					return err
				}

				err = pprof.StartCPUProfile(f)
				if err != nil {
					return err
				}
			}

			return nil
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if cpuprofile != "" {
				logrus.Info("stopping cpu profile")
				pprof.StopCPUProfile()
			}
			return nil
		},
	}

	root.PersistentFlags().BoolVar(&debug, "debug", false, "enable verbose output")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "do not output to console; use return code to determine success/failure")
	root.PersistentFlags().StringVarP(&cpuprofile, "cpuprofile", "p", "", "activate cpu profiling via pprof and write to file")

	return root

}
