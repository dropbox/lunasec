package controller

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"

	"github.com/Joker/jade"
	"github.com/pkg/errors"
	"github.com/refinery-labs/loq/model"
	"github.com/refinery-labs/loq/service"
	"github.com/refinery-labs/loq/util"
	"go.uber.org/config"
	"go.uber.org/zap"
)

type secureFrameController struct {
	SecureFrameControllerConfig
	logger   *zap.Logger
	indexTpl *template.Template
}

type SecureFrameControllerConfig struct {
	ViewsPath string          `yaml:"views_path"`
	CdnConfig model.CDNConfig `yaml:"cdn_config"`
}

type SecureFrameController interface {
	Frame(w http.ResponseWriter, r *http.Request)
}

func NewSecureFrameController(
	logger *zap.Logger,
	provider config.Provider,
) (controller SecureFrameController, err error) {
	var controllerConfig SecureFrameControllerConfig
	err = provider.Get("secure_frame_controller").Populate(&controllerConfig)
	if err != nil {
		return
	}

	jadeTpl, err := jade.ParseFile(getView(controllerConfig.ViewsPath, "index"))
	if err != nil {
		err = errors.Wrap(err, "unable to parse jade template file")
		return
	}

	indexTpl, err := template.New("html").Parse(jadeTpl)
	if err != nil {
		err = errors.Wrap(err, "unable to create html template")
		return
	}

	controller = &secureFrameController{
		SecureFrameControllerConfig: controllerConfig,
		logger:                      logger,
		indexTpl:                    indexTpl,
	}
	return
}

func (s *secureFrameController) Frame(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)

	referer := r.Header.Get("referer")

	if referer == "" {
		util.RespondError(w, http.StatusBadRequest, errors.New("missing origin for request"))
		return
	}

	query := r.URL.Query()

	nonce := query.Get("n")

	if nonce == "" {
		util.RespondError(w, http.StatusBadRequest, errors.New("missing unique id for request"))
		return
	}

	scriptURL := url.URL{
		Scheme: s.CdnConfig.Protocol,
		Host:   s.CdnConfig.Host,
		Path:   s.CdnConfig.MainScript,
	}

	styleURL := url.URL{
		Scheme: s.CdnConfig.Protocol,
		Host:   s.CdnConfig.Host,
		Path:   s.CdnConfig.MainStyle,
	}

	templateVars := model.FrameVars{
		CSPNonce:      service.Nonce(r.Context()),
		RequestOrigin: referer,
		RequestNonce:  nonce,
		ScriptUrl:     scriptURL.String(),
		StyleUrl:      styleURL.String(),
	}

	w.Header().Set("Content-Type", "text/html")
	err = s.indexTpl.Execute(w, templateVars)
	if err != nil {
		s.logger.Error("error returning website", zap.Error(err))
		util.RespondError(w, http.StatusBadRequest, errors.New("error returning website"))
	}
}

func getView(viewsPath, view string) string {
	return path.Join(viewsPath, fmt.Sprintf("%s.pug", view))
}