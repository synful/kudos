package main

import (
	"fmt"

	"github.com/joshlf/kudos/lib/build"
	"github.com/joshlf/kudos/lib/config"
	"github.com/joshlf/kudos/lib/dev"
	"github.com/joshlf/kudos/lib/kudos"
	"github.com/joshlf/kudos/lib/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var cmdMain = &cobra.Command{
	Use:   "kudos",
	Short: "kudos is a simple grading system",
	Long: `kudos is a simple grading system made out of love and frustration by m, ezr,
and jliebowf`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var verboseFlag bool
var quietFlag bool
var debugFlag bool
var configFlag string
var courseFlag string

// globalFlags contains flags which multiple
// different subcommands may use, but which
// the main command will not use.
//
// Subcommands should add these to themselves
// with addGlobalFlagsTo or addAllGlobalFlagsTo
var globalFlags = pflag.NewFlagSet("common", pflag.ContinueOnError)

// Make sure initGlobalFlags is called
// before any init functions are run
var _ = initGlobalFlags()

// Return a dummy value to enable the above line
func initGlobalFlags() struct{} {
	globalFlags.StringVarP(&configFlag, "config", "", config.DefaultGlobalConfigFile, "location of the global config file")
	globalFlags.StringVarP(&courseFlag, "course", "c", "", "course")
	return struct{}{}
}

// Add the named flags from globalFlags
// to the given flag set
func addGlobalFlagsTo(fset *pflag.FlagSet, flags ...string) {
	for _, fname := range flags {
		f := globalFlags.Lookup(fname)
		if f == nil {
			panic(fmt.Errorf("internal error: unkown named flag: %v", fname))
		}
		fset.AddFlag(f)
	}
}

// Add all flags in globalFlags
// to the given flag set
func addAllGlobalFlagsTo(fset *pflag.FlagSet) {
	globalFlags.VisitAll(func(f *pflag.Flag) { fset.AddFlag(f) })
}

func main() {
	// These flags are used by all subcommands,
	// but are also used by the main command
	// itself (so define them directly on
	// cmdMain rather than putting them in
	// globalFlags)
	cmdMain.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "be more verbose than normal; overrides --quiet")
	cmdMain.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "be more quiet than normal")
	cmdMain.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "print internal debugging information; implies --verbose")
	if build.DebugMode {
		cmdMain.DebugFlags()
	}
	cmdMain.Execute()
}

func getContext() *kudos.Context {
	c := &kudos.Context{}
	c.Logger = log.NewLogger()

	// If we're in debug mode, leave
	// debug logging on
	if !build.DebugMode {
		if debugFlag {
			c.SetLevel(log.Debug)
		} else if verboseFlag {
			c.SetLevel(log.Verbose)
		} else if quietFlag {
			c.SetLevel(log.Warn)
		}
	} else {
		if verboseFlag {
			c.Debug.Println("debug mode enabled; ignoring --verbose flag")
		}
		if quietFlag {
			c.Debug.Println("debug mode enabled; ignoring --quiet flag")
		}
	}

	return c
}

func addGlobalConfig(c *kudos.Context) {
	gcPath := config.DefaultGlobalConfigFile
	if globalFlags.Lookup("config").Changed {
		gcPath = configFlag
	}
	gc, err := kudos.ParseGlobalConfigFile(gcPath)
	if err != nil {
		c.Error.Printf("could not read global config: %v\n", err)
		dev.Fail()
	}
	c.GlobalConfig = gc
}

// implies addGlobalConfig(c)
func addCourse(c *kudos.Context) {
	addGlobalConfig(c)
	if !globalFlags.Lookup("course").Changed {
		c.Error.Println("no course provided; please specify one using the --course flag")
		dev.Fail()
	}
	code := courseFlag
	if err := kudos.ValidateCode(code); err != nil {
		c.Error.Printf("bad course code: %v\n", err)
		dev.Fail()
	}
	// since it was validated, code != ""
	// (which means we can safely assign
	// it to c.CourseCode, which requires
	// that if it is set, it is not "")
	c.CourseCode = code
}

// implies addCourse(c)
func addCourseConfig(c *kudos.Context) {
	addCourse(c)

	root := c.CourseRoot()
	course, err := kudos.ParseCourseFileValidateRoot(root)
	if err != nil {
		c.Error.Printf("could not read course config: %v\n", err)
		dev.Fail()
	}
	c.Course = course
}
