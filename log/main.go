package log

import (
	"fmt"
	stdLog "log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger() {
	const (
		DefaultFileMod os.FileMode = 0600
	)
	usr, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("Die on retrieving user info")
	}
	fileName := filepath.Join(os.TempDir(), fmt.Sprintf("term-%s.log", usr.Username))
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, DefaultFileMod)
	if err != nil {
		panic(err)
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"

	stdLog.SetFlags(stdLog.Lshortfile)
	stdLog.SetOutput(log.Output(zerolog.ConsoleWriter{Out: file}))

	stdLog.Printf("logger file init : %s", fileName)
}
