package streamlayer

import "github.com/rs/zerolog"

type streamlayerLogger struct {
	zerolog.Logger
}

func (stl *streamlayerLogger) Noticef(format string, v ...interface{}) {
	stl.Info().Msgf(format, v...)
}

func (stl *streamlayerLogger) Warnf(format string, v ...interface{}) {
	stl.Warn().Msgf(format, v...)
}

func (stl *streamlayerLogger) Fatalf(format string, v ...interface{}) {
	stl.Fatal().Msgf(format, v...)
}

func (stl *streamlayerLogger) Errorf(format string, v ...interface{}) {
	stl.Error().Msgf(format, v...)
}

func (stl *streamlayerLogger) Debugf(format string, v ...interface{}) {
	stl.Debug().Msgf(format, v...)
}

func (stl *streamlayerLogger) Tracef(format string, v ...interface{}) {
	stl.Trace().Msgf(format, v...)
}
