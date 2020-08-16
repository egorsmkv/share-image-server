/*

MIT License

Copyright (c) 2020 Yehor Smoliakov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package main

import (
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/fogleman/gg"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/sirupsen/logrus"
)

const (
	height = 630
	width  = 1200

	logoPath = "images/logo.png"

	sourceFontFile   = "fonts/FiraSans-ExtraBold.ttf"
	titleFontFile    = "fonts/FiraSans-Bold.ttf"
	categoryFontFile = "fonts/FiraSans-Medium.ttf"

	sourceWidth          = 700
	sourceHeight         = 200
	sourceNameSizePoints = 60

	logoWidth       = 100
	logoTopMargin   = 30
	logoRightMargin = 50

	topMargin  = 50
	leftMargin = 50

	titleBoxWidth     = (width / 2) + (width / 3)
	titleBoxHeight    = 250
	titleHeight       = titleBoxHeight + 30*2
	titleLeftMargin   = 30
	titleBottomMargin = (titleHeight / 2) - 20

	categoryNameSizePoints = 30
	titleNameSizePoints    = 55

	categoryBottomMargin = 100
)

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	logger.Level = logrus.DebugLevel

	logger.Info("Share Image Server is running")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		param := func(name string, isHex bool) string {
			val := r.URL.Query().Get(name)

			if isHex {
				return "#" + val
			}

			return val
		}

		disableLogo := param("disableLogo", false) == "yes"

		sourceName := param("sourceName", false)
		sourceNameColorHex := param("sourceNameColorHex", true)
		backgroundColorHex := param("backgroundColorHex", true)
		title := param("title", false)
		titleBackgroundColorHex := param("titleBackgroundColorHex", true)
		titleColorHex := param("titleColorHex", true)
		category := param("category", false)
		categoryColorHex := param("categoryColorHex", true)

		sourceNameColor, err := colorful.Hex(sourceNameColorHex)
		if err != nil {
			logger.WithError(err).WithField("sourceNameColorHex", sourceNameColorHex).Error("cannot create sourceNameColorHex")
			w.Write([]byte("cannot create sourceNameColorHex"))
			return
		}

		backgroundColor, err := colorful.Hex(backgroundColorHex)
		if err != nil {
			logger.WithError(err).WithField("backgroundColorHex", backgroundColorHex).Error("cannot create backgroundColorHex")
			w.Write([]byte("cannot create backgroundColorHex"))
			return
		}

		titleBackgroundColor, err := colorful.Hex(titleBackgroundColorHex)
		if err != nil {
			logger.WithError(err).WithField("titleBackgroundColorHex", titleBackgroundColorHex).Error("cannot create titleBackgroundColorHex")
			w.Write([]byte("cannot create titleBackgroundColorHex"))
			return
		}

		titleColor, err := colorful.Hex(titleColorHex)
		if err != nil {
			logger.WithError(err).WithField("titleColorHex", titleColorHex).Error("cannot create titleColorHex")
			w.Write([]byte("cannot create titleColorHex"))
			return
		}

		categoryColor, err := colorful.Hex(categoryColorHex)
		if err != nil {
			logger.WithError(err).WithField("categoryColorHex", categoryColorHex).Error("cannot create categoryColorHex")
			w.Write([]byte("cannot create categoryColorHex"))
			return
		}

		// Render

		// Create the main image
		dc := gg.NewContext(width, height)
		dc.DrawRectangle(0, 0, width, height)
		dc.SetColor(backgroundColor)
		dc.Fill()

		// Add the title
		dcSource := gg.NewContext(sourceWidth, sourceHeight)
		dcSource.SetColor(sourceNameColor)
		dcSource.Fill()
		if err := dcSource.LoadFontFace(sourceFontFile, sourceNameSizePoints); err != nil {
			logger.WithError(err).WithField("font", sourceFontFile).Error("cannot load font")
			w.Write([]byte("cannot load font"))
			return
		}
		dcSource.DrawStringAnchored(sourceName, 1, 1, 0, 1)

		// Add a box for the title
		titleBoxSource := gg.NewContext(titleBoxWidth, titleBoxHeight)
		titleBoxSource.DrawRoundedRectangle(0, 0, titleBoxWidth, titleBoxHeight, 10)
		titleBoxSource.SetColor(titleBackgroundColor)
		titleBoxSource.Fill()

		// Add the title
		titleSource := gg.NewContext(titleBoxWidth-20, titleHeight)
		titleSource.SetColor(titleColor)
		if err := titleSource.LoadFontFace(titleFontFile, titleNameSizePoints); err != nil {
			logger.WithError(err).WithField("font", titleFontFile).Error("cannot load font")
			w.Write([]byte("cannot load font"))
			return
		}
		titleSourceWidth := float64(titleBoxWidth - titleLeftMargin)
		titleSource.DrawStringWrapped(title, titleLeftMargin+460, titleBoxHeight-titleBottomMargin, 0.5, 0.5, titleSourceWidth, 1.5, gg.AlignLeft)

		// Add the title to the box
		titleBoxSource.DrawImage(titleSource.Image(), 10, 10)

		// Add the category
		categorySource := gg.NewContext(width, height)
		categorySource.SetColor(categoryColor)
		//categorySource.Fill()
		if err := categorySource.LoadFontFace(categoryFontFile, categoryNameSizePoints); err != nil {
			logger.WithError(err).WithField("font", categoryFontFile).Error("cannot load font")
			w.Write([]byte("cannot load font"))
			return
		}
		categorySourceWidth := float64(width - leftMargin)
		categorySource.DrawStringWrapped(category, 0, height-categoryBottomMargin, 0, 1, categorySourceWidth, 1.5, gg.AlignLeft)

		// Add the source to the main image
		dc.DrawImage(dcSource.Image(), leftMargin, topMargin)

		// Add the logo to the main image
		if !disableLogo {
			dcLogo, err := gg.LoadPNG(logoPath)
			if err != nil {
				logger.WithError(err).Error("cannot load logo")
				w.Write([]byte("cannot load logo"))
				return
			}
			dc.DrawImage(dcLogo, width-logoWidth-logoRightMargin, logoTopMargin)
		}

		// Add the box with its title to the main image
		dc.DrawImage(titleBoxSource.Image(), leftMargin, 200)

		// Add the category to the main image
		dc.DrawImage(categorySource.Image(), leftMargin, topMargin)

		// Create a PNG image and send it as the HTTP response
		err = png.Encode(w, dc.Image())
		if err != nil {
			logger.WithError(err).Error("cannot create image")
			w.Write([]byte("cannot create image"))
			return
		}

		logger.
			WithField("sourceName", sourceName).
			WithField("backgroundColorHex", backgroundColorHex).
			WithField("sourceNameColorHex", sourceNameColorHex).
			WithField("title", title).
			WithField("titleBackgroundColorHex", titleBackgroundColorHex).
			WithField("titleColorHex", titleColorHex).
			WithField("category", category).
			WithField("categoryColorHex", categoryColorHex).
			Debug("Provided parameters")
	})

	if err := http.ListenAndServe(":3000", r); err != nil {
		logger.WithError(err).Fatal("Cannot run the server")
	}
}

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

type StructuredLogger struct {
	Logger *logrus.Logger
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	logFields := logrus.Fields{}

	logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)

	entry.Logger.Infoln("request started")

	return entry
}

type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Infoln("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}
