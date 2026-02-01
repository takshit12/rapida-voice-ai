package internal_openai_callers

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"

	internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
	internal_caller_metrics "github.com/rapidaai/api/integration-api/internal/caller/metrics"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	protos "github.com/rapidaai/protos"
)

type largeLanguageCaller struct {
	OpenAI
}


func NewLargeLanguageCaller(logger commons.Logger, credential *protos.Credential) internal_callers.LargeLanguageCaller {
	return &largeLanguageCaller{
		OpenAI: openAI(logger, credential),
	}
}

func (llc *largeLanguageCaller) ChatCompletionOptions(
	opts *internal_callers.ChatCompletionOptions,
) openai.ChatCompletionNewParams {
	options := openai.ChatCompletionNewParams{}
	if len(opts.ToolDefinitions) > 0 {
		fns := make([]openai.ChatCompletionToolParam, len(opts.ToolDefinitions))
		for idx, tl := range opts.ToolDefinitions {
			switch tl.Type {
			case "tool":
			case "function":
				fn := tl.Function
				if fn != nil {
					funcDef := openai.FunctionDefinitionParam{
						Name: fn.Name,
					}
					if fn.Description != "" {
						funcDef.Description = openai.String(fn.Description)
					}
					// Always set parameters with valid JSON schema format
					if fn.Parameters != nil {
						funcDef.Parameters = fn.Parameters.ToMap()
					} else {
						// Default empty parameters with properties field for valid schema
						funcDef.Parameters = map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						}
					}
					fns[idx] = openai.ChatCompletionToolParam{
						Function: funcDef,
					}
				}
			}
		}
		options.Tools = fns
	}

	for key, value := range opts.ModelParameter {
		switch key {
		case "model.name":
			if modelName, err := utils.AnyToString(value); err == nil {
				options.Model = modelName
			}
		case "model.user":
			if user, err := utils.AnyToString(value); err == nil {
				options.User = openai.String(user)
			}
		case "model.reasoning_effort":
			if re, err := utils.AnyToString(value); err == nil {
				options.ReasoningEffort = shared.ReasoningEffort(re)
			}
		case "model.seed":
			if seed, err := utils.AnyToInt64(value); err == nil {
				options.Seed = openai.Int(seed)
			}
		case "model.service_tier":
			if st, err := utils.AnyToString(value); err == nil {
				options.ServiceTier = openai.ChatCompletionNewParamsServiceTier(st)
			}
		case "model.top_logprobs":
			if tl, err := utils.AnyToInt64(value); err == nil {
				options.TopLogprobs = openai.Int(tl)
			}
		case "model.metadata":
			format, _ := utils.AnyToString(value)
			var mtd map[string]string
			if err := json.Unmarshal([]byte(format), &mtd); err == nil {
				options.Metadata = shared.Metadata(mtd)
			}
		case "model.frequency_penalty":
			if fp, err := utils.AnyToFloat64(value); err == nil && fp != 0 {
				options.FrequencyPenalty = openai.Float(fp)
			}
		case "model.temperature":
			if temp, err := utils.AnyToFloat64(value); err == nil {
				options.Temperature = openai.Float(temp)
			}
		case "model.top_p":
			if topP, err := utils.AnyToFloat64(value); err == nil {
				options.TopP = openai.Float(topP)
			}
		case "model.presence_penalty":
			if pp, err := utils.AnyToFloat64(value); err == nil && pp != 0 {
				options.PresencePenalty = openai.Float(pp)
			}
		case "model.max_completion_tokens":
			if maxTokens, err := utils.AnyToInt64(value); err == nil {
				options.MaxCompletionTokens = openai.Int(maxTokens)
			}
		case "model.stop":
			if stopStr, err := utils.AnyToString(value); err == nil {
				for _, stopper := range strings.Split(stopStr, ",") {
					if strings.TrimSpace(stopper) != "" {
						options.Stop.OfStringArray = append(options.Stop.OfStringArray, stopper)
					}
				}
			}
		case "model.tool_choice":
			if choice, err := utils.AnyToString(value); err == nil {
				switch choice {
				case "auto":
					options.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
						OfAuto: openai.String("auto"),
					}
				case "required":
					options.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
						OfAuto: openai.String("required"),
					}
				case "none":
					options.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
						OfAuto: openai.String("none"),
					}
				// Empty or unknown values: don't set tool_choice, let the API use its default
				}
			}
		case "model.response_format":
			if format, err := utils.AnyToJSON(value); err == nil {
				switch format["type"].(string) {
				case "json_object":
					options.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
						OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
					}
				case "text":
					options.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{}
				case "json_schema":
					if schemaData, ok := format["json_schema"].(map[string]interface{}); ok {
						jsonSchemaParam := shared.ResponseFormatJSONSchemaJSONSchemaParam{}
						jsonData, err := json.Marshal(schemaData)
						if err == nil {
							json.Unmarshal(jsonData, &jsonSchemaParam)
						}
						options.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
							OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
								JSONSchema: jsonSchemaParam,
							},
						}
					}
				}
			}
		}
	}
	// Don't send tool_choice when no tools are defined — sending tool_choice
	// without tools causes GPT OSS 120B (and potentially other models) to
	// return empty responses with prompt_tokens: 0.
	if len(options.Tools) == 0 {
		options.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{}
	}
	// GPT OSS models: return a completely fresh params struct with ONLY the
	// fields that the working Groq curl example uses. The OpenAI SDK's apijson
	// serializer may include extra null fields (tool_choice, response_format,
	// audio, etc.) from union/struct types with embedded paramUnion metadata,
	// even when we don't set them. GPT-OSS on Groq may not tolerate these.
	if strings.HasPrefix(options.Model, "openai/gpt-oss") {
		cleanOptions := openai.ChatCompletionNewParams{
			Model:               options.Model,
			Temperature:         openai.Float(1),
			TopP:                openai.Float(1),
			MaxCompletionTokens: openai.Int(8192),
			ReasoningEffort:     shared.ReasoningEffort("medium"),
		}
		// Preserve tools if defined
		if len(options.Tools) > 0 {
			cleanOptions.Tools = options.Tools
			cleanOptions.ToolChoice = options.ToolChoice
		}
		return cleanOptions
	}
	return options
}

func (llc *largeLanguageCaller) GetChatCompletion(
	ctx context.Context,
	allMessages []*protos.Message,
	options *internal_callers.ChatCompletionOptions,
) (*types.Message, types.Metrics, error) {
	metrics := internal_caller_metrics.NewMetricBuilder(options.RequestId)
	metrics.OnStart()

	client, err := llc.GetClient()
	if err != nil {
		llc.logger.Errorf("chat completion unable to get client for openai %v", err)
		return nil, metrics.OnFailure().Build(), err
	}

	// message and options
	llmRequest := llc.ChatCompletionOptions(options)
	llmRequest.Messages = llc.BuildHistory(allMessages)

	// prehook
	options.PreHook(utils.ToJson(llmRequest))

	// chat complitions
	resp, err := client.Chat.Completions.New(ctx, llmRequest)
	if err != nil {
		llc.logger.Errorf("chat completion failed to get response from openai %v", err)
		options.PostHook(map[string]interface{}{
			"error":  err,
			"result": resp,
		}, metrics.OnFailure().Build())
		return nil, metrics.OnFailure().Build(), err
	}

	message := types.Message{
		Contents: make([]*types.Content, 0),
	}
	metrics.OnSuccess()
	metrics.OnAddMetrics(llc.GetComplitionUsages(resp.Usage)...)
	// all the usages into the metrics

	for _, choice := range resp.Choices {
		message.Role = string(choice.Message.Role)
		switch choice.FinishReason {
		case "length", "content_filter":
		case "stop":
			message.Contents = append(message.Contents, &types.Content{
				ContentType:   commons.TEXT_CONTENT.String(),
				ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
				Content:       []byte(choice.Message.Content),
			})
		case "function_call", "tool_calls":
			if choice.Message.ToolCalls != nil {
				for _, tool := range choice.Message.ToolCalls {
					if tool.Type == "function" {
						newToolCall := &types.ToolCall{
							Id:   &tool.ID,
							Type: utils.Ptr(string(tool.Type)),
							Function: &types.FunctionCall{
								Name:      utils.Ptr(tool.Function.Name),
								Arguments: utils.Ptr(tool.Function.Arguments),
							},
						}
						if message.ToolCalls == nil {
							message.ToolCalls = make([]*types.ToolCall, 0)
						}
						message.ToolCalls = append(message.ToolCalls, newToolCall)
					}
				}
			}
		}
	}

	options.PostHook(map[string]interface{}{
		"result": resp,
	}, metrics.OnSuccess().Build())
	return &message, metrics.Build(), nil
}

func (llc *largeLanguageCaller) StreamChatCompletion(
	ctx context.Context,
	allMessages []*protos.Message,
	options *internal_callers.ChatCompletionOptions,
	onStream func(types.Message) error,
	onMetrics func(*types.Message, types.Metrics) error,
	onError func(err error),
) error {
	start := time.Now()
	metrics := internal_caller_metrics.NewMetricBuilder(options.RequestId)
	metrics.OnStart()

	completionsOptions := llc.ChatCompletionOptions(options)

	// GPT-OSS models: bypass the OpenAI SDK entirely and use raw HTTP/1.1.
	// The SDK works for other models but GPT-OSS on Groq returns empty responses
	// through the SDK despite identical JSON bodies and headers. This eliminates
	// all SDK transport-level variables (HTTP/2, connection pooling, etc.).
	modelName := ""
	if mn, ok := options.ModelParameter["model.name"]; ok {
		modelName, _ = utils.AnyToString(mn)
	}
	if strings.HasPrefix(modelName, "openai/gpt-oss") || strings.HasPrefix(completionsOptions.Model, "openai/gpt-oss") {
		credentials := llc.credential()
		apiKey, _ := credentials[API_KEY].(string)
		baseURL := DEFAULT_URL
		if url, ok := credentials[API_URL]; ok && url != nil {
			if urlStr, ok := url.(string); ok && urlStr != "" {
				baseURL = urlStr
			}
		}
		// Build plain messages from proto
		var plainMessages []map[string]string
		for _, msg := range allMessages {
			role := msg.GetRole()
			content := types.OnlyStringProtoContent(msg.GetContents())
			if content != "" {
				plainMessages = append(plainMessages, map[string]string{
					"role":    role,
					"content": content,
				})
			}
		}
		options.PreHook(map[string]interface{}{
			"model":    completionsOptions.Model,
			"messages": plainMessages,
			"method":   "rawStreamGPTOSS",
		})
		return llc.rawStreamGPTOSS(ctx, apiKey, baseURL, plainMessages,
			completionsOptions.Model, onStream, onMetrics, onError, metrics, options)
	}

	client, err := llc.GetClient()
	if err != nil {
		llc.logger.Errorf("chat completion unable to get client for openai: %v", err)
		onError(err)
		onMetrics(nil, metrics.OnFailure().Build())
		return err
	}

	completionsOptions.Messages = llc.BuildHistory(allMessages)

	llc.logger.Infof("stream request: model=%q messages=%d tools=%d maxTokens=%v maxCompletionTokens=%v temperature=%v reasoning_effort=%q",
		completionsOptions.Model, len(completionsOptions.Messages), len(completionsOptions.Tools),
		completionsOptions.MaxTokens, completionsOptions.MaxCompletionTokens, completionsOptions.Temperature,
		completionsOptions.ReasoningEffort)
	// Log the full request JSON to verify serialization (excluding messages for brevity)
	reqCopy := completionsOptions
	reqCopy.Messages = nil
	if reqJSON, err := json.Marshal(reqCopy); err == nil {
		llc.logger.Infof("stream request JSON (no messages): %s", string(reqJSON))
	}
	options.PreHook(utils.ToJson(completionsOptions))
	llc.logger.Benchmark("Openai.llm.GetChatCompletion.llmRequestPrepare", time.Since(start))

	// Get streaming response
	resp := client.Chat.Completions.NewStreaming(ctx, completionsOptions)
	if resp.Err() != nil {
		llc.logger.Errorf("Failed to get chat completions stream: %v", resp.Err())
		options.PostHook(map[string]interface{}{
			"result": utils.ToJson(resp),
			"error":  resp.Err(),
		}, metrics.Build())
		onMetrics(nil, metrics.OnFailure().Build())
		onError(resp.Err())
		return resp.Err()
	}
	defer resp.Close()
	completeMsg := types.Message{
		Role:      "assistant",
		Contents:  make([]*types.Content, 0),
		ToolCalls: make([]*types.ToolCall, 0),
	}

	chunkCount := 0
	accumulate := openai.ChatCompletionAccumulator{}
	for resp.Next() {
		chatCompletions := resp.Current()
		accumulate.AddChunk(chatCompletions)
		chunkCount++

		// Log raw chunk JSON for debugging streaming issues (especially reasoning models)
		if rawJSON, err := json.Marshal(chatCompletions); err == nil {
			llc.logger.Infof("stream chunk #%d raw: %s", chunkCount, string(rawJSON))
		}
		for ci, choice := range chatCompletions.Choices {
			llc.logger.Debugf("stream chunk #%d choice[%d]: delta.content=%q delta.role=%q finish_reason=%q refusal=%q",
				chunkCount, ci, choice.Delta.Content, choice.Delta.Role, choice.FinishReason, choice.Delta.Refusal)
		}
		if len(chatCompletions.Choices) == 0 {
			llc.logger.Debugf("stream chunk #%d: no choices (usage update or empty chunk)", chunkCount)
		}

		// Process delta content BEFORE checking JustFinishedContent/JustFinishedToolCall.
		// Some models (e.g. Groq GPT OSS) send content and finish_reason in the same chunk.
		// If we check JustFinishedContent first, the delta from that chunk is never processed.
		deltaMsg := types.Message{
			Role:     "assistant",
			Contents: make([]*types.Content, 0),
		}

		for i, choice := range chatCompletions.Choices {
			content := choice.Delta.Content
			if content != "" {
				// Update complete message
				if len(completeMsg.Contents) <= i {
					completeMsg.Contents = append(completeMsg.Contents, &types.Content{
						ContentType:   commons.TEXT_CONTENT.String(),
						ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
						Content:       []byte(content),
					})
				} else {
					completeMsg.Contents[i].Content = append(completeMsg.Contents[i].Content, []byte(content)...)
				}
				// Update delta message
				deltaMsg.Contents = append(deltaMsg.Contents, &types.Content{
					ContentType:   commons.TEXT_CONTENT.String(),
					ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
					Content:       []byte(content),
				})
			}
		}

		// Stream content if there are changes
		if len(deltaMsg.Contents) > 0 {
			if err := onStream(deltaMsg); err != nil {
				llc.logger.Errorf("Error sending stream data: %v", err)
				return err
			}
		}

		if _, ok := accumulate.JustFinishedContent(); ok {
			llc.logger.Infof("stream: JustFinishedContent after %d chunks, completeMsg.Contents=%d, accumulated=%q",
				chunkCount, len(completeMsg.Contents), accumulate.Choices[0].Message.Content)
			// If completeMsg is empty but accumulator has content, use accumulated content
			if len(completeMsg.Contents) == 0 && len(accumulate.Choices) > 0 && accumulate.Choices[0].Message.Content != "" {
				llc.logger.Infof("stream: JustFinishedContent using accumulated content as fallback (delta was empty)")
				completeMsg.Contents = append(completeMsg.Contents, &types.Content{
					ContentType:   commons.TEXT_CONTENT.String(),
					ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
					Content:       []byte(accumulate.Choices[0].Message.Content),
				})
				completeMsg.Role = string(accumulate.Choices[0].Message.Role)
				// Send content via onStream so it reaches the UI as LLMStreamPacket
				// (onMetrics sends Data+Metrics together, but processStream skips Data when metrics != nil)
				if err := onStream(completeMsg); err != nil {
					llc.logger.Errorf("Error sending accumulated stream data: %v", err)
					return err
				}
			}
			metrics.OnAddMetrics(llc.GetComplitionUsages(accumulate.Usage)...)
			metrics.OnSuccess()
			options.PostHook(map[string]interface{}{
				"result": utils.ToJson(accumulate),
			}, metrics.Build())
			onMetrics(&completeMsg, metrics.Build())
			return nil
		}

		if tool, ok := accumulate.JustFinishedToolCall(); ok {
			completeMsg.ToolCalls = append(completeMsg.ToolCalls, &types.ToolCall{
				Id: utils.Ptr(tool.ID),
				Function: &types.FunctionCall{
					Name:      utils.Ptr(tool.Name),
					Arguments: utils.Ptr(tool.Arguments),
				},
			})

			// Stream the complete message with tool calls
			if err := onStream(completeMsg); err != nil {
				llc.logger.Errorf("Error sending tool call data: %v", err)
				return err
			}

			metrics.OnAddMetrics(llc.GetComplitionUsages(accumulate.Usage)...)
			options.PostHook(map[string]interface{}{
				"result": utils.ToJson(accumulate),
			}, metrics.Build())
			onMetrics(&completeMsg, metrics.Build())
			return nil
		}
	}

	// Check for streaming errors
	if resp.Err() != nil {
		llc.logger.Errorf("Stream error after loop: %v", resp.Err())
		onError(resp.Err())
		onMetrics(nil, metrics.OnFailure().Build())
		return resp.Err()
	}

	// Fallback: if JustFinishedContent/JustFinishedToolCall never fired,
	// still send metrics so the caller gets proper completion
	llc.logger.Warnf("stream: loop exited without JustFinishedContent/JustFinishedToolCall after %d chunks, completeMsg.Contents=%d",
		chunkCount, len(completeMsg.Contents))
	if len(accumulate.Choices) > 0 {
		llc.logger.Warnf("stream: accumulated content=%q, finish_reason=%q",
			accumulate.Choices[0].Message.Content, accumulate.Choices[0].FinishReason)
		// If we have accumulated content but completeMsg is empty, use accumulated content
		if len(completeMsg.Contents) == 0 && accumulate.Choices[0].Message.Content != "" {
			llc.logger.Infof("stream: using accumulated content as fallback")
			completeMsg.Contents = append(completeMsg.Contents, &types.Content{
				ContentType:   commons.TEXT_CONTENT.String(),
				ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
				Content:       []byte(accumulate.Choices[0].Message.Content),
			})
			completeMsg.Role = string(accumulate.Choices[0].Message.Role)
		}
	}
	// Send content via onStream so it reaches the UI as LLMStreamPacket
	// (onMetrics sends Data+Metrics together, but processStream skips Data when metrics != nil)
	if len(completeMsg.Contents) > 0 {
		if err := onStream(completeMsg); err != nil {
			llc.logger.Errorf("Error sending fallback stream data: %v", err)
			return err
		}
	}
	metrics.OnAddMetrics(llc.GetComplitionUsages(accumulate.Usage)...)
	metrics.OnSuccess()
	options.PostHook(map[string]interface{}{
		"result": utils.ToJson(accumulate),
	}, metrics.Build())
	onMetrics(&completeMsg, metrics.Build())
	return nil
}

// rawStreamGPTOSS bypasses the OpenAI SDK entirely and makes a raw HTTP/1.1
// request to the Groq API for GPT-OSS models. This eliminates all SDK-related
// variables (HTTP/2, connection pooling, header injection, serialization).
func (llc *largeLanguageCaller) rawStreamGPTOSS(
	ctx context.Context,
	apiKey string,
	baseURL string,
	messages []map[string]string,
	model string,
	onStream func(types.Message) error,
	onMetrics func(*types.Message, types.Metrics) error,
	onError func(err error),
	metrics *internal_caller_metrics.MetricBuilder,
	options *internal_callers.ChatCompletionOptions,
) error {
	// Build request body exactly matching working curl from Groq playground
	reqBody := map[string]interface{}{
		"messages":              messages,
		"model":                 model,
		"temperature":           1,
		"max_completion_tokens": 8192,
		"top_p":                 1,
		"stream":                true,
		"reasoning_effort":      "medium",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		onError(fmt.Errorf("failed to marshal request: %w", err))
		onMetrics(nil, metrics.OnFailure().Build())
		return err
	}

	llc.logger.Infof("rawStreamGPTOSS: POST %s/chat/completions body=%s", baseURL, string(bodyBytes))

	// Create raw HTTP/1.1 request (matching curl behavior exactly)
	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		onError(fmt.Errorf("failed to create request: %w", err))
		onMetrics(nil, metrics.OnFailure().Build())
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Force HTTP/1.1 to match curl (Go defaults to HTTP/2 with TLS)
	transport := &http.Transport{
		ForceAttemptHTTP2: false,
		TLSNextProto:     make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	httpClient := &http.Client{Transport: transport}

	resp, err := httpClient.Do(req)
	if err != nil {
		llc.logger.Errorf("rawStreamGPTOSS: HTTP error: %v", err)
		onError(fmt.Errorf("HTTP request failed: %w", err))
		onMetrics(nil, metrics.OnFailure().Build())
		return err
	}
	defer resp.Body.Close()

	llc.logger.Infof("rawStreamGPTOSS: response status=%d proto=%s", resp.StatusCode, resp.Proto)
	for key, values := range resp.Header {
		llc.logger.Infof("rawStreamGPTOSS: response header %s: %s", key, values)
	}

	if resp.StatusCode != 200 {
		respBytes, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBytes))
		llc.logger.Errorf("rawStreamGPTOSS: %s", errMsg)
		onError(fmt.Errorf(errMsg))
		onMetrics(nil, metrics.OnFailure().Build())
		return fmt.Errorf(errMsg)
	}

	// Parse SSE response line by line
	scanner := bufio.NewScanner(resp.Body)
	// Increase scanner buffer for large chunks
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	completeMsg := types.Message{
		Role:     "assistant",
		Contents: make([]*types.Content, 0),
	}
	var fullContent strings.Builder
	chunkCount := 0
	var lastUsage struct {
		PromptTokens     int
		CompletionTokens int
		TotalTokens      int
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			llc.logger.Infof("rawStreamGPTOSS: received [DONE] after %d chunks", chunkCount)
			break
		}

		chunkCount++
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			llc.logger.Warnf("rawStreamGPTOSS: failed to parse chunk #%d: %v", chunkCount, err)
			continue
		}

		llc.logger.Infof("rawStreamGPTOSS: chunk #%d: %s", chunkCount, data)

		// Track usage
		if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
			lastUsage.PromptTokens = chunk.Usage.PromptTokens
			lastUsage.CompletionTokens = chunk.Usage.CompletionTokens
			lastUsage.TotalTokens = chunk.Usage.TotalTokens
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				deltaMsg := types.Message{
					Role: "assistant",
					Contents: []*types.Content{{
						ContentType:   commons.TEXT_CONTENT.String(),
						ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
						Content:       []byte(choice.Delta.Content),
					}},
				}
				if err := onStream(deltaMsg); err != nil {
					llc.logger.Errorf("rawStreamGPTOSS: error sending stream: %v", err)
					return err
				}
			}

			if choice.FinishReason != nil && *choice.FinishReason == "stop" {
				llc.logger.Infof("rawStreamGPTOSS: finish_reason=stop after %d chunks, content_len=%d", chunkCount, fullContent.Len())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		llc.logger.Errorf("rawStreamGPTOSS: scanner error: %v", err)
		onError(err)
		onMetrics(nil, metrics.OnFailure().Build())
		return err
	}

	// Build complete message
	llc.logger.Infof("rawStreamGPTOSS: stream complete. chunks=%d content_len=%d prompt_tokens=%d completion_tokens=%d",
		chunkCount, fullContent.Len(), lastUsage.PromptTokens, lastUsage.CompletionTokens)

	if fullContent.Len() > 0 {
		completeMsg.Contents = append(completeMsg.Contents, &types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(fullContent.String()),
		})
	}

	metrics.OnAddMetrics(
		&types.Metric{Name: type_enums.OUTPUT_TOKEN.String(), Value: fmt.Sprintf("%d", lastUsage.CompletionTokens), Description: "Output Token"},
		&types.Metric{Name: type_enums.INPUT_TOKEN.String(), Value: fmt.Sprintf("%d", lastUsage.PromptTokens), Description: "Input Token"},
		&types.Metric{Name: type_enums.TOTAL_TOKEN.String(), Value: fmt.Sprintf("%d", lastUsage.TotalTokens), Description: "Total Token"},
	)
	metrics.OnSuccess()
	options.PostHook(map[string]interface{}{
		"result":   fullContent.String(),
		"method":   "rawStreamGPTOSS",
		"chunks":   chunkCount,
		"usage":    lastUsage,
	}, metrics.Build())
	onMetrics(&completeMsg, metrics.Build())
	return nil
}

func (llc *largeLanguageCaller) BuildHistory(allMessages []*protos.Message) []openai.ChatCompletionMessageParamUnion {
	msg := make([]openai.ChatCompletionMessageParamUnion, 0)
	for _, cntn := range allMessages {
		switch cntn.GetRole() {
		case ChatRoleUser:
			// Check if all content is text-only (no images). If so, use simple
			// string format for broader compatibility (e.g. Groq GPT OSS models
			// may not support the content parts array format).
			allText := true
			var textParts []string
			for _, ct := range cntn.GetContents() {
				if ct.ContentType == commons.TEXT_CONTENT.String() {
					textParts = append(textParts, string(ct.GetContent()))
				} else {
					allText = false
					break
				}
			}
			if allText && len(textParts) > 0 {
				msg = append(msg, openai.UserMessage(strings.Join(textParts, "\n")))
			} else {
				var messageContent []openai.ChatCompletionContentPartUnionParam
				for _, ct := range cntn.GetContents() {
					switch ct.ContentType {
					case commons.TEXT_CONTENT.String():
						messageContent = append(messageContent, openai.ChatCompletionContentPartUnionParam{
							OfText: &openai.ChatCompletionContentPartTextParam{
								Text: string(ct.GetContent()),
							},
						})
					case commons.IMAGE_CONTENT.String():
						if ct.GetContentFormat() == commons.IMAGE_CONTENT_FORMAT_URL.String() {
							messageContent = append(messageContent, openai.ChatCompletionContentPartUnionParam{
								OfImageURL: &openai.ChatCompletionContentPartImageParam{
									ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
										URL: string(ct.GetContent()),
									},
								},
							})
						}
					default:
						llc.logger.Warnf("Unknown content type: %s", ct.ContentType)
					}
				}
				msg = append(msg, openai.UserMessage(messageContent))
			}
		case ChatRoleAssistant:
			txtContent := types.OnlyStringProtoContent(cntn.GetContents())
			toolCalls := cntn.GetToolCalls()
			assistantMessage := openai.ChatCompletionAssistantMessageParam{}
			if len(txtContent) > 0 || len(toolCalls) > 0 {
				if len(txtContent) > 0 {
					assistantMessage.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.String(txtContent),
					}
				}
				if len(toolCalls) > 0 {
					fctCall := make([]openai.ChatCompletionMessageToolCallParam, 0)
					for _, ttc := range toolCalls {
						fctCall = append(fctCall, openai.ChatCompletionMessageToolCallParam{
							ID: ttc.GetId(),
							Function: openai.ChatCompletionMessageToolCallFunctionParam{
								Name:      ttc.GetFunction().GetName(),
								Arguments: ttc.GetFunction().GetArguments(),
							},
						})
					}
					assistantMessage.ToolCalls = fctCall
				}
				msg = append(msg, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &assistantMessage,
				})
			}

		case ChatRoleSystem:
			txtContent := types.OnlyStringProtoContent(cntn.GetContents())
			if len(txtContent) > 0 {
				msg = append(msg, openai.SystemMessage(txtContent))
			}

		case ChatRoleTool:
			for _, tcl := range cntn.GetContents() {
				toolId := tcl.GetContentType()
				msg = append(msg, openai.ToolMessage(string(tcl.GetContent()), toolId))
			}
		}
	}
	return msg
}
