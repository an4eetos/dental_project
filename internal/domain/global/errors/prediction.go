package errors

var (
	ErrNoPredictionExamples = BaseError{
		Msg:  "no prediction examples configured",
		ErrC: "no_prediction_examples",
	}
	ErrPredictionFailed = BaseError{
		Msg:  "prediction failed",
		ErrC: "prediction_failed",
	}
)
