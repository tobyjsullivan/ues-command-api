package consumer

import (
    "log"
    "os"
)

var (
    logger *log.Logger
)

func init()  {
    logger = log.New(os.Stdout, "[consumer] ", 0)
}
