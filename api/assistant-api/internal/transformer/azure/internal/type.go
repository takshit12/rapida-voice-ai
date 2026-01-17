// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package azure_internal

// azureRecognizingResult represents the JSON structure for interim recognition results.
type AzureRecognizingResult struct {
	ID              string `json:"Id"`
	Text            string `json:"Text"`
	Offset          int64  `json:"Offset"`
	Duration        int64  `json:"Duration"`
	PrimaryLanguage struct {
		Language string `json:"Language"`
	} `json:"PrimaryLanguage"`
	Channel int `json:"Channel"`
}

// azureRecognizedResult represents the JSON structure for final recognition results.
type AzureRecognizedResult struct {
	ID                string `json:"Id"`
	RecognitionStatus string `json:"RecognitionStatus"`
	DisplayText       string `json:"DisplayText"`
	NBest             []struct {
		Confidence float64 `json:"Confidence"`
		Lexical    string  `json:"Lexical"`
		Display    string  `json:"Display"`
		Words      []struct {
			Word       string  `json:"Word"`
			Confidence float64 `json:"Confidence"`
		} `json:"Words"`
	} `json:"NBest"`
}
