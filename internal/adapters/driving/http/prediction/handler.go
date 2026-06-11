package prediction

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	domain "github.com/anuarkuanysh/dental_project/internal/domain/prediction"
	predictionuc "github.com/anuarkuanysh/dental_project/internal/usecase/prediction"
)

// Handler exposes the few-shot prediction endpoint.
type Handler struct {
	PredictUC *predictionuc.Predict
}

type predictRequest struct {
	Age                         string `json:"age" binding:"required"`
	PregnancyWeeks              string `json:"pregnancy_weeks" binding:"required"`
	KPUIndex                    string `json:"kpu_index" binding:"required"`
	BrushingPerDay              string `json:"brushing_per_day" binding:"required"`
	DentistVisitDuringPregnancy string `json:"dentist_visit_during_pregnancy" binding:"required"`
	ParentCaries                string `json:"parent_caries" binding:"required"`
	SalivaPH                    string `json:"saliva_ph" binding:"required"`
}

type predictResponse struct {
	ChildCariesProbability string `json:"child_caries_probability"`
	RiskGroup              string `json:"risk_group"`
	Action                 string `json:"action"`
	Recommendations        string `json:"recommendations"`
}

// Predict accepts survey inputs and returns the four generated outputs.
func (h *Handler) Predict(c *gin.Context) {
	var req predictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "all survey fields are required",
		})
		return
	}

	output, err := h.PredictUC.Execute(c.Request.Context(), toInputRow(req))
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, predictResponse{
		ChildCariesProbability: output[domain.KeyChildCariesProbability],
		RiskGroup:              output[domain.KeyRiskGroup],
		Action:                 output[domain.KeyAction],
		Recommendations:        output[domain.KeyRecommendations],
	})
}

func toInputRow(req predictRequest) domain.Row {
	return domain.Row{
		domain.KeyAge:                         strings.TrimSpace(req.Age),
		domain.KeyPregnancyWeeks:              strings.TrimSpace(req.PregnancyWeeks),
		domain.KeyKPUIndex:                    strings.TrimSpace(req.KPUIndex),
		domain.KeyBrushingPerDay:              strings.TrimSpace(req.BrushingPerDay),
		domain.KeyDentistVisitDuringPregnancy: strings.TrimSpace(req.DentistVisitDuringPregnancy),
		domain.KeyParentCaries:                strings.TrimSpace(req.ParentCaries),
		domain.KeySalivaPH:                    strings.TrimSpace(req.SalivaPH),
	}
}
