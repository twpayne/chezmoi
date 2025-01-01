package cmd

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Suppress log output from github.com/zricethezav/gitleaks/v8.
	// See https://github.com/gitleaks/gitleaks/issues/1684.
	log.Logger = zerolog.New(io.Discard)
}
