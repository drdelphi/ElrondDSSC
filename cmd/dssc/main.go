package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DrDelphi/ElrondDSSC/bot"
	"github.com/DrDelphi/ElrondDSSC/config"
	"github.com/DrDelphi/ElrondDSSC/db"
	"github.com/DrDelphi/ElrondDSSC/network"
	"github.com/DrDelphi/ElrondDSSC/utils"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/core/logging"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath      = "logs"
	logFilePrefix        = "elrond-dssc"
	logFileLifeSpanInSec = 86400
)

var (
	dsscHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`
	// configPathFlag defines a flag for the path of the application's configuration file
	configPathFlag = cli.StringFlag{
		Name:  "config-path",
		Usage: "The application will load its configuration parameters from this file",
		Value: utils.DefaultConfigPath,
	}
	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
	//logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
)

var log = logger.GetOrCreate("main")

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = dsscHelpTemplate
	app.Name = "Elrond Delegation System SC CLI App"
	app.Usage = "This app is a Telegram Bot for interacting with the Elrond DSSC"
	app.Flags = []cli.Flag{
		configPathFlag,
		logLevel,
		logSaveFile,
	}
	app.Version = "v0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "DrDelphi",
			Email: "drdelphi@gmail.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startApp(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startApp(ctx *cli.Context) error {
	var err error

	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	withLogFile := ctx.GlobalBool(logSaveFile.Name)
	var fileLogging factory.FileLoggingHandler
	if withLogFile {
		workingDir := getWorkingDir(log)
		fileLogging, err = logging.NewFileLogging(workingDir, defaultLogsPath, logFilePrefix)
		if err != nil {
			return fmt.Errorf("%w creating a log file", err)
		}

		err = fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(logFileLifeSpanInSec))
		if err != nil {
			return err
		}
	}

	log.Info("starting elrond dssc...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info("loading config...")

	configurationFileName := ctx.GlobalString(configPathFlag.Name)
	appConfig, err := config.NewConfig(configurationFileName)
	if err != nil {
		return err
	}

	log.Info("opening database...")

	database, err := db.NewDatabase(appConfig.DatabasePath)
	if err != nil {
		return err
	}

	log.Info("initializing network manager...")

	networkManager, err := network.NewNetworkManager(appConfig)
	if err != nil {
		return err
	}

	log.Info("creating Telegram bot instance...")

	tgBot, err := bot.NewBot(appConfig, database, networkManager)
	if err != nil {
		return err
	}

	tgBot.StartTasks()

	log.Info("application is now running...")

	mainLoop(sigs)

	log.Debug("closing elrond dssc...")
	if fileLogging != nil {
		err = fileLogging.Close()
		log.LogIfError(err)
	}

	return nil
}

func getWorkingDir(log logger.Logger) string {
	workingDir, err := os.Getwd()
	if err != nil {
		log.LogIfError(err)
		workingDir = ""
	}

	log.Trace("working directory", "path", workingDir)

	return workingDir
}

func mainLoop(stop chan os.Signal) {
	for {
		select {
		case <-stop:
			log.Info("terminating at user's signal...")
			return
		}
	}
}
