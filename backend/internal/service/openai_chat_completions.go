package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
	"github.com/gin-gonic/gin"
)

const openaiChatCompletionPath = "/chat/completions"

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatStreamState struct {
	id          string
	created     int64
	model       string
	sentRole    bool
	hasToolCall bool
	toolIndex   map[string]int
	toolNames   map[string]string
}

type toolCallState struct {
	index int
	id    string
	name  string
}

func (s *OpenAIGatewayService) ForwardChatCompletions(ctx context.Context, c *gin.Context, account *Account, body []byte) (*OpenAIForwardResult, error) {
	startTime := time.Now()

	var reqBody map[string]any
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return nil, fmt.Errorf("parse request: %w", err)
	}

	reqModel, _ := reqBody["model"].(string)
	reqStream, _ := reqBody["stream"].(bool)
	originalModel := reqModel
	promptCacheKey := ""
	if v, ok := reqBody["prompt_cache_key"].(string); ok {
		promptCacheKey = strings.TrimSpace(v)
	}

	includeUsage := chatStreamIncludeUsage(reqBody)
	isCodexCLI := openai.IsCodexCLIRequest(c.GetHeader("User-Agent"))

	bodyModified := false
	mappedModel := account.GetMappedModel(reqModel)
	if mappedModel != reqModel {
		reqBody["model"] = mappedModel
		bodyModified = true
	}

	if model, ok := reqBody["model"].(string); ok {
		normalizedModel := normalizeCodexModel(model)
		if normalizedModel != "" && normalizedModel != model {
			reqBody["model"] = normalizedModel
			mappedModel = normalizedModel
			bodyModified = true
		}
	}

	useResponses := account.Type == AccountTypeOAuth

	if useResponses {
		converted, err := convertChatCompletionsToResponses(reqBody)
		if err != nil {
			s.writeChatErrorResponse(c, http.StatusBadRequest, "invalid_request_error", err.Error())
			return nil, fmt.Errorf("convert request: %w", err)
		}
		reqBody = converted
		bodyModified = true
	}

	if effort, ok := reqBody["reasoning_effort"].(string); ok {
		if strings.TrimSpace(effort) != "" {
			reasoning, _ := reqBody["reasoning"].(map[string]any)
			if reasoning == nil {
				reasoning = map[string]any{}
			}
			if _, ok := reasoning["effort"]; !ok {
				reasoning["effort"] = effort
			}
			delete(reqBody, "reasoning_effort")
			reqBody["reasoning"] = reasoning
			bodyModified = true
		}
	}

	if reasoning, ok := reqBody["reasoning"].(map[string]any); ok {
		if effort, ok := reasoning["effort"].(string); ok && effort == "minimal" {
			reasoning["effort"] = "none"
			bodyModified = true
		}
	}

	if reqStream && !includeUsage {
		if ensureChatStreamUsage(reqBody) {
			bodyModified = true
		}
	}

	if bodyModified {
		var err error
		body, err = json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("serialize request body: %w", err)
		}
	}

	token, _, err := s.GetAccessToken(ctx, account)
	if err != nil {
		return nil, err
	}

	var upstreamReq *http.Request
	if useResponses {
		upstreamReq, err = s.buildUpstreamRequest(ctx, c, account, body, token, reqStream, promptCacheKey, isCodexCLI)
	} else {
		upstreamReq, err = s.buildUpstreamChatRequest(ctx, c, account, body, token)
	}
	if err != nil {
		return nil, err
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	if c != nil {
		c.Set(OpsUpstreamRequestBodyKey, string(body))
	}

	resp, err := s.httpUpstream.Do(upstreamReq, proxyURL, account.ID, account.Concurrency)
	if err != nil {
		safeErr := sanitizeUpstreamErrorMessage(err.Error())
		setOpsUpstreamError(c, 0, safeErr, "")
		appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
			Platform:           account.Platform,
			AccountID:          account.ID,
			AccountName:        account.Name,
			UpstreamStatusCode: 0,
			Kind:               "request_error",
			Message:            safeErr,
		})
		s.writeChatErrorResponse(c, http.StatusBadGateway, "upstream_error", "Upstream request failed")
		return nil, fmt.Errorf("upstream request failed: %s", safeErr)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		if s.shouldFailoverUpstreamError(resp.StatusCode) {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
			_ = resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(respBody))

			upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(respBody))
			upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
			upstreamDetail := ""
			if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
				maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
				if maxBytes <= 0 {
					maxBytes = 2048
				}
				upstreamDetail = truncateString(string(respBody), maxBytes)
			}
			appendOpsUpstreamError(c, OpsUpstreamErrorEvent{
				Platform:           account.Platform,
				AccountID:          account.ID,
				AccountName:        account.Name,
				UpstreamStatusCode: resp.StatusCode,
				UpstreamRequestID:  resp.Header.Get("x-request-id"),
				Kind:               "failover",
				Message:            upstreamMsg,
				Detail:             upstreamDetail,
			})

			s.handleFailoverSideEffects(ctx, resp, account)
			return nil, &UpstreamFailoverError{StatusCode: resp.StatusCode}
		}
		return s.handleChatErrorResponse(ctx, resp, c, account)
	}

	var usage *OpenAIUsage
	var firstTokenMs *int

	if reqStream {
		if useResponses {
			streamResult, err := s.handleChatCompletionsFromResponsesStream(ctx, resp, c, account, startTime, originalModel, mappedModel, includeUsage)
			if err != nil {
				return nil, err
			}
			usage = streamResult.usage
			firstTokenMs = streamResult.firstTokenMs
		} else {
			streamResult, err := s.handleChatCompletionsStream(ctx, resp, c, account, startTime, originalModel, mappedModel, includeUsage)
			if err != nil {
				return nil, err
			}
			usage = streamResult.usage
			firstTokenMs = streamResult.firstTokenMs
		}
	} else {
		if useResponses {
			usage, err = s.handleChatCompletionsFromResponses(ctx, resp, c, account, originalModel, mappedModel)
		} else {
			usage, err = s.handleChatCompletionsNonStreaming(ctx, resp, c, account, originalModel, mappedModel)
		}
		if err != nil {
			return nil, err
		}
	}

	if account.Type == AccountTypeOAuth {
		if snapshot := ParseCodexRateLimitHeaders(resp.Header); snapshot != nil {
			s.updateCodexUsageSnapshot(ctx, account.ID, snapshot)
		}
	}

	if usage == nil {
		usage = &OpenAIUsage{}
	}

	return &OpenAIForwardResult{
		RequestID:    resp.Header.Get("x-request-id"),
		Usage:        *usage,
		Model:        originalModel,
		Stream:       reqStream,
		Duration:     time.Since(startTime),
		FirstTokenMs: firstTokenMs,
	}, nil
}

func (s *OpenAIGatewayService) buildUpstreamChatRequest(ctx context.Context, c *gin.Context, account *Account, body []byte, token string) (*http.Request, error) {
	var targetURL string
	baseURL := account.GetOpenAIBaseURL()
	if baseURL == "" {
		targetURL = "https://api.openai.com/v1" + openaiChatCompletionPath
	} else {
		validatedURL, err := s.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, err
		}
		validatedURL = strings.TrimRight(validatedURL, "/")
		if !strings.HasSuffix(validatedURL, "/v1") {
			validatedURL = validatedURL + "/v1"
		}
		targetURL = validatedURL + openaiChatCompletionPath
	}

	req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("authorization", "Bearer "+token)

	for key, values := range c.Request.Header {
		lowerKey := strings.ToLower(key)
		if openaiAllowedHeaders[lowerKey] {
			for _, v := range values {
				req.Header.Add(key, v)
			}
		}
	}

	customUA := account.GetOpenAIUserAgent()
	if customUA != "" {
		req.Header.Set("user-agent", customUA)
	}

	if req.Header.Get("content-type") == "" {
		req.Header.Set("content-type", "application/json")
	}

	return req, nil
}

func (s *OpenAIGatewayService) handleChatCompletionsStream(ctx context.Context, resp *http.Response, c *gin.Context, account *Account, startTime time.Time, originalModel, mappedModel string, includeUsage bool) (*openaiStreamingResult, error) {
	if s.cfg != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.cfg.Security.ResponseHeaders)
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	if v := resp.Header.Get("x-request-id"); v != "" {
		c.Header("x-request-id", v)
	}

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	usage := &OpenAIUsage{}
	var firstTokenMs *int

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	bodyClosed := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = resp.Body.Close()
		case <-bodyClosed:
		}
	}()
	defer close(bodyClosed)

	type scanEvent struct {
		line string
		err  error
	}

	events := make(chan scanEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev scanEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}

	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func() {
		defer close(events)
		for scanner.Scan() {
			atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			if !sendEvent(scanEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			_ = sendEvent(scanEvent{err: err})
		}
	}()
	defer close(done)

	streamInterval := time.Duration(0)
	if s.cfg != nil && s.cfg.Gateway.StreamDataIntervalTimeout > 0 {
		streamInterval = time.Duration(s.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
	}

	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	keepaliveInterval := time.Duration(0)
	if s.cfg != nil && s.cfg.Gateway.StreamKeepaliveInterval > 0 {
		keepaliveInterval = time.Duration(s.cfg.Gateway.StreamKeepaliveInterval) * time.Second
	}
	var keepaliveTicker *time.Ticker
	if keepaliveInterval > 0 {
		keepaliveTicker = time.NewTicker(keepaliveInterval)
		defer keepaliveTicker.Stop()
	}
	var keepaliveCh <-chan time.Time
	if keepaliveTicker != nil {
		keepaliveCh = keepaliveTicker.C
	}
	lastDataAt := time.Now()

	errorEventSent := false
	sendErrorEvent := func(reason string) {
		if errorEventSent {
			return
		}
		errorEventSent = true
		_, _ = fmt.Fprintf(w, "event: error\ndata: {\"error\":{\"message\":\"%s\"}}\n\n", reason)
		flusher.Flush()
	}

	needModelReplace := originalModel != mappedModel

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, nil
			}
			if ev.err != nil {
				if errors.Is(ev.err, bufio.ErrTooLong) {
					log.Printf("SSE line too long: account=%d max_size=%d error=%v", account.ID, maxLineSize, ev.err)
					sendErrorEvent("response_too_large")
					return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, ev.err
				}
				sendErrorEvent("stream_read_error")
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream read error: %w", ev.err)
			}

			line := ev.line
			lastDataAt = time.Now()

			if openaiSSEDataRe.MatchString(line) {
				data := openaiSSEDataRe.ReplaceAllString(line, "")
				if data == "" {
					continue
				}

				if needModelReplace {
					data = replaceChatStreamModel(data, mappedModel, originalModel)
					line = "data: " + data
				}

				if correctedData, corrected := s.toolCorrector.CorrectToolCallsInSSEData(data); corrected {
					data = correctedData
					line = "data: " + correctedData
				}

				if err := updateChatUsageFromStreamChunk(data, usage); err == nil {
					if !includeUsage {
						if suppressed, ok := removeChatUsageFromChunk(data); ok {
							if suppressed == "" {
								continue
							}
							data = suppressed
							line = "data: " + data
						}
					}
				}

				if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
					sendErrorEvent("write_failed")
					return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
				}
				flusher.Flush()

				if firstTokenMs == nil && data != "" && data != "[DONE]" {
					ms := int(time.Since(startTime).Milliseconds())
					firstTokenMs = &ms
				}
			} else {
				if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
					sendErrorEvent("write_failed")
					return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
				}
				flusher.Flush()
			}

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			log.Printf("Stream data interval timeout: account=%d model=%s interval=%s", account.ID, originalModel, streamInterval)
			if s.rateLimitService != nil {
				s.rateLimitService.HandleStreamTimeout(ctx, account, originalModel)
			}
			sendErrorEvent("stream_timeout")
			return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream data interval timeout")

		case <-keepaliveCh:
			if time.Since(lastDataAt) < keepaliveInterval {
				continue
			}
			if _, err := fmt.Fprint(w, ":\n\n"); err != nil {
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
			}
			flusher.Flush()
		}
	}
}

func (s *OpenAIGatewayService) handleChatCompletionsNonStreaming(ctx context.Context, resp *http.Response, c *gin.Context, account *Account, originalModel, mappedModel string) (*OpenAIUsage, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	usage := &OpenAIUsage{}
	if parsed, ok := parseChatCompletionUsage(body); ok {
		usage = parsed
	}

	if originalModel != mappedModel {
		body = replaceChatCompletionModel(body, mappedModel, originalModel)
	}

	corrected := s.correctToolCallsInResponseBody(body)
	if len(corrected) > 0 {
		body = corrected
	}

	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.cfg.Security.ResponseHeaders)

	contentType := "application/json"
	if s.cfg != nil && !s.cfg.Security.ResponseHeaders.Enabled {
		if upstreamType := resp.Header.Get("Content-Type"); upstreamType != "" {
			contentType = upstreamType
		}
	}

	c.Data(resp.StatusCode, contentType, body)
	return usage, nil
}

func (s *OpenAIGatewayService) handleChatCompletionsFromResponses(ctx context.Context, resp *http.Response, c *gin.Context, account *Account, originalModel, mappedModel string) (*OpenAIUsage, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if isEventStreamResponse(resp.Header) || bytes.Contains(body, []byte("data:")) {
		if finalResponse, ok := extractCodexFinalResponse(string(body)); ok {
			body = finalResponse
		}
	}

	chatBody, usage, err := convertResponsesToChatCompletions(body, originalModel, mappedModel)
	if err != nil {
		return nil, err
	}

	responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.cfg.Security.ResponseHeaders)

	contentType := "application/json"
	if s.cfg != nil && !s.cfg.Security.ResponseHeaders.Enabled {
		if upstreamType := resp.Header.Get("Content-Type"); upstreamType != "" {
			contentType = upstreamType
		}
	}

	c.Data(resp.StatusCode, contentType, chatBody)
	return usage, nil
}

func (s *OpenAIGatewayService) handleChatCompletionsFromResponsesStream(ctx context.Context, resp *http.Response, c *gin.Context, account *Account, startTime time.Time, originalModel, mappedModel string, includeUsage bool) (*openaiStreamingResult, error) {
	if s.cfg != nil {
		responseheaders.WriteFilteredHeaders(c.Writer.Header(), resp.Header, s.cfg.Security.ResponseHeaders)
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	if v := resp.Header.Get("x-request-id"); v != "" {
		c.Header("x-request-id", v)
	}

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	usage := &OpenAIUsage{}
	var firstTokenMs *int

	scanner := bufio.NewScanner(resp.Body)
	maxLineSize := defaultMaxLineSize
	if s.cfg != nil && s.cfg.Gateway.MaxLineSize > 0 {
		maxLineSize = s.cfg.Gateway.MaxLineSize
	}
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	bodyClosed := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = resp.Body.Close()
		case <-bodyClosed:
		}
	}()
	defer close(bodyClosed)

	type scanEvent struct {
		line string
		err  error
	}

	events := make(chan scanEvent, 16)
	done := make(chan struct{})
	sendEvent := func(ev scanEvent) bool {
		select {
		case events <- ev:
			return true
		case <-done:
			return false
		}
	}

	var lastReadAt int64
	atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
	go func() {
		defer close(events)
		for scanner.Scan() {
			atomic.StoreInt64(&lastReadAt, time.Now().UnixNano())
			if !sendEvent(scanEvent{line: scanner.Text()}) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			_ = sendEvent(scanEvent{err: err})
		}
	}()
	defer close(done)

	streamInterval := time.Duration(0)
	if s.cfg != nil && s.cfg.Gateway.StreamDataIntervalTimeout > 0 {
		streamInterval = time.Duration(s.cfg.Gateway.StreamDataIntervalTimeout) * time.Second
	}

	var intervalTicker *time.Ticker
	if streamInterval > 0 {
		intervalTicker = time.NewTicker(streamInterval)
		defer intervalTicker.Stop()
	}
	var intervalCh <-chan time.Time
	if intervalTicker != nil {
		intervalCh = intervalTicker.C
	}

	keepaliveInterval := time.Duration(0)
	if s.cfg != nil && s.cfg.Gateway.StreamKeepaliveInterval > 0 {
		keepaliveInterval = time.Duration(s.cfg.Gateway.StreamKeepaliveInterval) * time.Second
	}
	var keepaliveTicker *time.Ticker
	if keepaliveInterval > 0 {
		keepaliveTicker = time.NewTicker(keepaliveInterval)
		defer keepaliveTicker.Stop()
	}
	var keepaliveCh <-chan time.Time
	if keepaliveTicker != nil {
		keepaliveCh = keepaliveTicker.C
	}
	lastDataAt := time.Now()

	state := chatStreamState{
		model:     originalModel,
		toolIndex: map[string]int{},
		toolNames: map[string]string{},
	}

	errorEventSent := false
	sendErrorEvent := func(reason string) {
		if errorEventSent {
			return
		}
		errorEventSent = true
		_, _ = fmt.Fprintf(w, "event: error\ndata: {\"error\":{\"message\":\"%s\"}}\n\n", reason)
		flusher.Flush()
	}

	sendChunk := func(delta map[string]any, finishReason any) error {
		if state.id == "" {
			state.id = "chatcmpl-unknown"
		}
		if state.created == 0 {
			state.created = time.Now().Unix()
		}
		chunk := map[string]any{
			"id":      state.id,
			"object":  "chat.completion.chunk",
			"created": state.created,
			"model":   originalModel,
			"choices": []any{
				map[string]any{
					"index":         0,
					"delta":         delta,
					"finish_reason": finishReason,
				},
			},
		}
		payload, err := json.Marshal(chunk)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	ensureRole := func() error {
		if state.sentRole {
			return nil
		}
		state.sentRole = true
		return sendChunk(map[string]any{"role": "assistant"}, nil)
	}

	finalize := func(finishReason string) error {
		if finishReason == "" {
			finishReason = "stop"
		}
		if err := sendChunk(map[string]any{}, finishReason); err != nil {
			return err
		}
		if includeUsage {
			usagePayload := buildChatUsagePayload(usage)
			finalChunk := map[string]any{
				"id":      state.id,
				"object":  "chat.completion.chunk",
				"created": state.created,
				"model":   originalModel,
				"choices": []any{},
				"usage":   usagePayload,
			}
			payload, err := json.Marshal(finalChunk)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
				return err
			}
			flusher.Flush()
		}
		if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, nil
			}
			if ev.err != nil {
				if errors.Is(ev.err, bufio.ErrTooLong) {
					log.Printf("SSE line too long: account=%d max_size=%d error=%v", account.ID, maxLineSize, ev.err)
					sendErrorEvent("response_too_large")
					return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, ev.err
				}
				sendErrorEvent("stream_read_error")
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream read error: %w", ev.err)
			}

			line := ev.line
			lastDataAt = time.Now()

			if !openaiSSEDataRe.MatchString(line) {
				continue
			}

			data := openaiSSEDataRe.ReplaceAllString(line, "")
			if data == "" || data == "[DONE]" {
				continue
			}

			var payload map[string]any
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				continue
			}

			eventType, _ := payload["type"].(string)

			switch eventType {
			case "response.created", "response.in_progress":
				if response, ok := payload["response"].(map[string]any); ok {
					if id, ok := response["id"].(string); ok && id != "" {
						state.id = id
					}
					if created, ok := response["created_at"].(float64); ok && created > 0 {
						state.created = int64(created)
					}
					if model, ok := response["model"].(string); ok && model != "" {
						state.model = model
					}
				}

			case "response.output_item.added":
				item, _ := payload["item"].(map[string]any)
				itemType, _ := item["type"].(string)
				if itemType == "message" {
					if err := ensureRole(); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
				}
				if itemType == "function_call" || itemType == "tool_call" {
					toolState := buildToolState(item, &state)
					if toolState != nil {
						if err := ensureRole(); err != nil {
							sendErrorEvent("write_failed")
							return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
						}
						toolDelta := map[string]any{
							"index": toolState.index,
							"id":    toolState.id,
							"type":  "function",
							"function": map[string]any{
								"name": toolState.name,
							},
						}
						if err := sendChunk(map[string]any{"tool_calls": []any{toolDelta}}, nil); err != nil {
							sendErrorEvent("write_failed")
							return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
						}
						state.hasToolCall = true
						if firstTokenMs == nil {
							ms := int(time.Since(startTime).Milliseconds())
							firstTokenMs = &ms
						}
					}
				}

			case "response.output_text.delta":
				delta, _ := payload["delta"].(string)
				if delta != "" {
					if err := ensureRole(); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					if err := sendChunk(map[string]any{"content": delta}, nil); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					if firstTokenMs == nil {
						ms := int(time.Since(startTime).Milliseconds())
						firstTokenMs = &ms
					}
				}

			case "response.refusal.delta":
				delta, _ := payload["delta"].(string)
				if delta != "" {
					if err := ensureRole(); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					if err := sendChunk(map[string]any{"refusal": delta}, nil); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					if firstTokenMs == nil {
						ms := int(time.Since(startTime).Milliseconds())
						firstTokenMs = &ms
					}
				}

			case "response.reasoning_text.delta", "response.reasoning_summary_text.delta":
				delta, _ := payload["delta"].(string)
				if delta != "" {
					if err := ensureRole(); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					if err := sendChunk(map[string]any{"reasoning_content": delta}, nil); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					if firstTokenMs == nil {
						ms := int(time.Since(startTime).Milliseconds())
						firstTokenMs = &ms
					}
				}

			case "response.function_call_arguments.delta":
				itemID, _ := payload["item_id"].(string)
				delta, _ := payload["delta"].(string)
				if itemID != "" && delta != "" {
					callID := itemID
					if err := ensureRole(); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					index := getToolIndex(&state, itemID)
					toolDelta := map[string]any{
						"index": index,
						"id":    callID,
						"type":  "function",
						"function": map[string]any{
							"arguments": delta,
						},
					}
					if name := state.toolNames[itemID]; name != "" {
						if fn, ok := toolDelta["function"].(map[string]any); ok {
							fn["name"] = name
						}
					}
					if err := sendChunk(map[string]any{"tool_calls": []any{toolDelta}}, nil); err != nil {
						sendErrorEvent("write_failed")
						return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
					}
					state.hasToolCall = true
					if firstTokenMs == nil {
						ms := int(time.Since(startTime).Milliseconds())
						firstTokenMs = &ms
					}
				}

			case "response.completed", "response.incomplete", "response.failed":
				s.parseSSEUsage(data, usage)
				finishReason := resolveChatFinishReason(payload, &state)
				if err := finalize(finishReason); err != nil {
					return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
				}
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, nil

			case "error":
				if errMsg := extractErrorMessage(payload); errMsg != "" {
					sendErrorEvent(errMsg)
					return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("upstream error: %s", errMsg)
				}
			}

		case <-intervalCh:
			lastRead := time.Unix(0, atomic.LoadInt64(&lastReadAt))
			if time.Since(lastRead) < streamInterval {
				continue
			}
			log.Printf("Stream data interval timeout: account=%d model=%s interval=%s", account.ID, originalModel, streamInterval)
			if s.rateLimitService != nil {
				s.rateLimitService.HandleStreamTimeout(ctx, account, originalModel)
			}
			sendErrorEvent("stream_timeout")
			return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, fmt.Errorf("stream data interval timeout")

		case <-keepaliveCh:
			if time.Since(lastDataAt) < keepaliveInterval {
				continue
			}
			if _, err := fmt.Fprint(w, ":\n\n"); err != nil {
				return &openaiStreamingResult{usage: usage, firstTokenMs: firstTokenMs}, err
			}
			flusher.Flush()
		}
	}
}

func convertChatCompletionsToResponses(reqBody map[string]any) (map[string]any, error) {
	if reqBody == nil {
		return nil, fmt.Errorf("empty request")
	}

	messages, ok := reqBody["messages"].([]any)
	if !ok {
		return nil, fmt.Errorf("messages must be an array")
	}

	filteredMessages := make([]any, 0, len(messages))
	var instructionsParts []string
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			continue
		}
		role, _ := msgMap["role"].(string)
		role = strings.TrimSpace(role)
		if role == "system" {
			if content := strings.TrimSpace(extractChatContentAsString(msgMap["content"])); content != "" {
				instructionsParts = append(instructionsParts, content)
			}
			continue
		}
		filteredMessages = append(filteredMessages, msg)
	}

	inputItems, err := convertChatMessagesToInput(filteredMessages)
	if err != nil {
		return nil, err
	}

	converted := make(map[string]any, len(reqBody))
	for key, value := range reqBody {
		converted[key] = value
	}

	converted["input"] = inputItems
	delete(converted, "messages")
	if _, ok := converted["instructions"]; !ok {
		if len(instructionsParts) > 0 {
			converted["instructions"] = strings.Join(instructionsParts, "\n")
		}
	}
	if _, ok := converted["store"]; !ok {
		converted["store"] = false
	}
	if v, ok := converted["stream"].(bool); !ok || !v {
		converted["stream"] = true
	}

	if _, ok := converted["tools"]; !ok {
		if functions, ok := converted["functions"].([]any); ok {
			converted["tools"] = convertChatFunctionsToTools(functions)
		}
	}
	delete(converted, "functions")

	if _, ok := converted["tool_choice"]; !ok {
		if fc, exists := converted["function_call"]; exists {
			converted["tool_choice"] = convertFunctionCallToToolChoice(fc)
		}
	}
	delete(converted, "function_call")

	if _, ok := converted["text"]; !ok {
		if responseFormat, ok := converted["response_format"]; ok {
			converted["text"] = map[string]any{"format": responseFormat}
		}
	}
	delete(converted, "response_format")

	if _, ok := converted["max_output_tokens"]; !ok {
		if v, ok := converted["max_completion_tokens"]; ok {
			converted["max_output_tokens"] = v
		} else if v, ok := converted["max_tokens"]; ok {
			converted["max_output_tokens"] = v
		}
	}
	delete(converted, "max_completion_tokens")
	delete(converted, "max_tokens")

	if normalizeCodexTools(converted) {
		// Ensure tools are in responses format.
	}

	return converted, nil
}

func convertChatMessagesToInput(messages []any) ([]any, error) {
	input := make([]any, 0, len(messages))
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			continue
		}
		role, _ := msgMap["role"].(string)
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}

		if role == "tool" || role == "function" {
			output := map[string]any{
				"type":   "function_call_output",
				"output": extractChatContentAsString(msgMap["content"]),
			}
			callID := ""
			if value, ok := msgMap["tool_call_id"].(string); ok {
				callID = strings.TrimSpace(value)
			}
			name := ""
			if value, ok := msgMap["name"].(string); ok {
				name = strings.TrimSpace(value)
			}
			if callID == "" && name == "" {
				return nil, fmt.Errorf("tool messages require tool_call_id or name")
			}
			if callID == "" {
				callID = name
			}
			output["call_id"] = callID
			if name != "" {
				output["name"] = name
			}
			input = append(input, output)
			continue
		}

		if role == "assistant" {
			if toolCalls, ok := msgMap["tool_calls"].([]any); ok {
				for _, call := range toolCalls {
					toolItem := convertChatToolCallToInput(call)
					if toolItem != nil {
						input = append(input, toolItem)
					}
				}
			}
			if functionCall, ok := msgMap["function_call"].(map[string]any); ok {
				toolItem := convertChatFunctionCallToInput(functionCall)
				if toolItem != nil {
					input = append(input, toolItem)
				}
			}
		}

		contentParts, ok := convertChatContentToInputParts(msgMap["content"])
		if !ok {
			continue
		}

		messageItem := map[string]any{
			"role":    role,
			"content": contentParts,
		}
		if name, ok := msgMap["name"].(string); ok && strings.TrimSpace(name) != "" {
			messageItem["name"] = name
		}
		input = append(input, messageItem)
	}
	return input, nil
}

func convertChatContentToInputParts(content any) ([]any, bool) {
	switch v := content.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, false
		}
		return []any{map[string]any{"type": "input_text", "text": v}}, true
	case []any:
		parts := make([]any, 0, len(v))
		for _, part := range v {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			partType, _ := partMap["type"].(string)
			partType = strings.TrimSpace(partType)
			switch partType {
			case "text":
				if text, ok := partMap["text"].(string); ok {
					parts = append(parts, map[string]any{"type": "input_text", "text": text})
				}
			case "image_url":
				if imageURL, ok := partMap["image_url"]; ok {
					parts = append(parts, map[string]any{"type": "input_image", "image_url": imageURL})
				}
			default:
				parts = append(parts, partMap)
			}
		}
		if len(parts) == 0 {
			return nil, false
		}
		return parts, true
	default:
		return nil, false
	}
}

func extractChatContentAsString(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var builder strings.Builder
		for _, part := range v {
			partMap, ok := part.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := partMap["text"].(string); ok {
				builder.WriteString(text)
			}
		}
		return builder.String()
	default:
		return ""
	}
}

func convertChatToolCallToInput(call any) map[string]any {
	callMap, ok := call.(map[string]any)
	if !ok {
		return nil
	}
	callID, _ := callMap["id"].(string)
	function, _ := callMap["function"].(map[string]any)
	name, _ := function["name"].(string)
	arguments := function["arguments"]

	item := map[string]any{
		"type": "function_call",
	}
	if strings.TrimSpace(callID) != "" {
		item["call_id"] = callID
		item["id"] = callID
	}
	if strings.TrimSpace(name) != "" {
		item["name"] = name
	}
	if arguments != nil {
		item["arguments"] = arguments
	}
	return item
}

func convertChatFunctionCallToInput(functionCall map[string]any) map[string]any {
	name, _ := functionCall["name"].(string)
	arguments := functionCall["arguments"]
	if strings.TrimSpace(name) == "" && arguments == nil {
		return nil
	}
	item := map[string]any{
		"type": "function_call",
	}
	if strings.TrimSpace(name) != "" {
		item["name"] = name
		item["call_id"] = name
		item["id"] = name
	}
	if arguments != nil {
		item["arguments"] = arguments
	}
	return item
}

func convertChatFunctionsToTools(functions []any) []any {
	tools := make([]any, 0, len(functions))
	for _, fn := range functions {
		fnMap, ok := fn.(map[string]any)
		if !ok {
			continue
		}
		tools = append(tools, map[string]any{
			"type":     "function",
			"function": fnMap,
		})
	}
	return tools
}

func convertFunctionCallToToolChoice(functionCall any) any {
	switch value := functionCall.(type) {
	case string:
		return value
	case map[string]any:
		name, _ := value["name"].(string)
		if strings.TrimSpace(name) == "" {
			return value
		}
		return map[string]any{
			"type": "function",
			"name": name,
		}
	default:
		return functionCall
	}
}

func chatStreamIncludeUsage(reqBody map[string]any) bool {
	if reqBody == nil {
		return false
	}
	streamOptions, ok := reqBody["stream_options"].(map[string]any)
	if !ok || streamOptions == nil {
		return false
	}
	includeUsage, _ := streamOptions["include_usage"].(bool)
	return includeUsage
}

func ensureChatStreamUsage(reqBody map[string]any) bool {
	streamOptions, ok := reqBody["stream_options"].(map[string]any)
	if !ok || streamOptions == nil {
		streamOptions = map[string]any{}
		reqBody["stream_options"] = streamOptions
	}
	if includeUsage, ok := streamOptions["include_usage"].(bool); ok && includeUsage {
		return false
	}
	streamOptions["include_usage"] = true
	return true
}

func parseChatCompletionUsage(body []byte) (*OpenAIUsage, bool) {
	var payload struct {
		Usage chatUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, false
	}
	if payload.Usage.PromptTokens == 0 && payload.Usage.CompletionTokens == 0 && payload.Usage.TotalTokens == 0 {
		return nil, false
	}
	return &OpenAIUsage{
		InputTokens:  payload.Usage.PromptTokens,
		OutputTokens: payload.Usage.CompletionTokens,
	}, true
}

func updateChatUsageFromStreamChunk(data string, usage *OpenAIUsage) error {
	if data == "" {
		return fmt.Errorf("empty")
	}
	var payload struct {
		Usage *chatUsage `json:"usage"`
	}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return err
	}
	if payload.Usage == nil {
		return fmt.Errorf("no usage")
	}
	usage.InputTokens = payload.Usage.PromptTokens
	usage.OutputTokens = payload.Usage.CompletionTokens
	return nil
}

func removeChatUsageFromChunk(data string) (string, bool) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return "", false
	}
	if _, ok := payload["usage"]; !ok {
		return "", false
	}
	delete(payload, "usage")
	if choices, ok := payload["choices"].([]any); ok && len(choices) == 0 {
		return "", true
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return "", false
	}
	return string(out), true
}

func replaceChatCompletionModel(body []byte, fromModel, toModel string) []byte {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return body
	}
	if model, ok := resp["model"].(string); ok && model == fromModel {
		resp["model"] = toModel
	}
	newBody, err := json.Marshal(resp)
	if err != nil {
		return body
	}
	return newBody
}

func replaceChatStreamModel(data string, fromModel, toModel string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return data
	}
	if model, ok := payload["model"].(string); ok && model == fromModel {
		payload["model"] = toModel
	}
	out, err := json.Marshal(payload)
	if err != nil {
		return data
	}
	return string(out)
}

func convertResponsesToChatCompletions(body []byte, originalModel, mappedModel string) ([]byte, *OpenAIUsage, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, fmt.Errorf("parse response: %w", err)
	}

	if originalModel != mappedModel {
		if model, ok := resp["model"].(string); ok && model == mappedModel {
			resp["model"] = originalModel
		}
	}

	id, _ := resp["id"].(string)
	created := int64(0)
	if v, ok := resp["created_at"].(float64); ok {
		created = int64(v)
	}
	model, _ := resp["model"].(string)

	message, toolCalls, reasoningContent, refusalText, annotations := extractChatMessageFromResponses(resp)

	finishReason := resolveChatFinishReasonFromResponse(resp, len(toolCalls) > 0)

	if message == nil {
		message = map[string]any{"role": "assistant", "content": ""}
	}
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
	}
	if reasoningContent != "" {
		message["reasoning_content"] = reasoningContent
	}
	if refusalText != "" {
		message["refusal"] = refusalText
	}
	if len(annotations) > 0 {
		message["annotations"] = annotations
	}

	usage := &OpenAIUsage{}
	if usageMap, ok := resp["usage"].(map[string]any); ok {
		usage.InputTokens = intFromAny(usageMap["input_tokens"])
		usage.OutputTokens = intFromAny(usageMap["output_tokens"])
		if details, ok := usageMap["input_tokens_details"].(map[string]any); ok {
			usage.CacheReadInputTokens = intFromAny(details["cached_tokens"])
		}
	}

	chatResp := map[string]any{
		"id":      id,
		"object":  "chat.completion",
		"created": created,
		"model":   model,
		"choices": []any{
			map[string]any{
				"index":         0,
				"message":       message,
				"finish_reason": finishReason,
			},
		},
		"usage": buildChatUsagePayload(usage),
	}

	payload, err := json.Marshal(chatResp)
	if err != nil {
		return nil, nil, err
	}
	return payload, usage, nil
}

func extractChatMessageFromResponses(resp map[string]any) (map[string]any, []any, string, string, []any) {
	output, ok := resp["output"].([]any)
	if !ok {
		return nil, nil, "", "", nil
	}

	var contentParts []string
	var reasoningParts []string
	var refusalParts []string
	var annotations []any
	var toolCalls []any
	message := map[string]any{"role": "assistant"}

	for _, item := range output {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := itemMap["type"].(string)
		switch itemType {
		case "message":
			if role, ok := itemMap["role"].(string); ok && role != "" {
				message["role"] = role
			}
			if content, ok := itemMap["content"].([]any); ok {
				for _, part := range content {
					partMap, ok := part.(map[string]any)
					if !ok {
						continue
					}
					partType, _ := partMap["type"].(string)
					switch partType {
					case "output_text":
						if text, ok := partMap["text"].(string); ok {
							contentParts = append(contentParts, text)
						}
						if ann, ok := partMap["annotations"].([]any); ok {
							annotations = append(annotations, ann...)
						}
					case "reasoning_text", "reasoning_summary_text", "output_reasoning":
						if text, ok := partMap["text"].(string); ok {
							reasoningParts = append(reasoningParts, text)
						}
					case "refusal":
						if text, ok := partMap["text"].(string); ok {
							refusalParts = append(refusalParts, text)
						}
					}
				}
			}
		case "function_call", "tool_call":
			toolCall := convertResponsesToolCallToChat(itemMap)
			if toolCall != nil {
				toolCalls = append(toolCalls, toolCall)
			}
		}
	}

	if len(contentParts) > 0 {
		message["content"] = strings.Join(contentParts, "")
	} else {
		message["content"] = nil
	}
	return message, toolCalls, strings.Join(reasoningParts, ""), strings.Join(refusalParts, ""), annotations
}

func convertResponsesToolCallToChat(item map[string]any) map[string]any {
	callID, _ := item["call_id"].(string)
	if callID == "" {
		if id, ok := item["id"].(string); ok {
			callID = id
		}
	}
	name, _ := item["name"].(string)
	if corrected, ok := CorrectToolName(name); ok {
		name = corrected
	}
	tool := map[string]any{
		"id":   callID,
		"type": "function",
		"function": map[string]any{
			"name": name,
		},
	}
	if args, ok := item["arguments"]; ok && args != nil {
		if fn, ok := tool["function"].(map[string]any); ok {
			fn["arguments"] = args
		}
	}
	return tool
}

func resolveChatFinishReasonFromResponse(resp map[string]any, hasToolCalls bool) string {
	if hasToolCalls {
		return "tool_calls"
	}
	if status, ok := resp["status"].(string); ok {
		if status == "incomplete" {
			return "length"
		}
	}
	return "stop"
}

func resolveChatFinishReason(payload map[string]any, state *chatStreamState) string {
	if state != nil && state.hasToolCall {
		return "tool_calls"
	}
	if response, ok := payload["response"].(map[string]any); ok {
		if status, ok := response["status"].(string); ok {
			if status == "incomplete" {
				return "length"
			}
		}
	}
	return "stop"
}

func buildChatUsagePayload(usage *OpenAIUsage) map[string]any {
	if usage == nil {
		return nil
	}
	total := usage.InputTokens + usage.OutputTokens
	payload := map[string]any{
		"prompt_tokens":     usage.InputTokens,
		"completion_tokens": usage.OutputTokens,
		"total_tokens":      total,
		"prompt_tokens_details": map[string]any{
			"cached_tokens": usage.CacheReadInputTokens,
		},
		"completion_tokens_details": map[string]any{
			"reasoning_tokens": 0,
		},
	}
	return payload
}

func buildToolState(item map[string]any, state *chatStreamState) *toolCallState {
	if item == nil || state == nil {
		return nil
	}
	callID, _ := item["call_id"].(string)
	if callID == "" {
		if id, ok := item["id"].(string); ok {
			callID = id
		}
	}
	name, _ := item["name"].(string)
	if corrected, ok := CorrectToolName(name); ok {
		name = corrected
	}
	if callID == "" {
		return nil
	}
	index := getToolIndex(state, callID)
	state.toolNames[callID] = name
	return &toolCallState{index: index, id: callID, name: name}
}

func getToolIndex(state *chatStreamState, key string) int {
	if state == nil {
		return 0
	}
	if idx, ok := state.toolIndex[key]; ok {
		return idx
	}
	idx := len(state.toolIndex)
	state.toolIndex[key] = idx
	return idx
}

func extractErrorMessage(payload map[string]any) string {
	if errMap, ok := payload["error"].(map[string]any); ok {
		if msg, ok := errMap["message"].(string); ok {
			return sanitizeUpstreamErrorMessage(msg)
		}
	}
	return ""
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		parsed, _ := strconv.Atoi(v)
		return parsed
	default:
		return 0
	}
}

func (s *OpenAIGatewayService) handleChatErrorResponse(ctx context.Context, resp *http.Response, c *gin.Context, account *Account) (*OpenAIForwardResult, error) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	upstreamMsg := strings.TrimSpace(extractUpstreamErrorMessage(body))
	upstreamMsg = sanitizeUpstreamErrorMessage(upstreamMsg)
	upstreamDetail := ""
	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		maxBytes := s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes
		if maxBytes <= 0 {
			maxBytes = 2048
		}
		upstreamDetail = truncateString(string(body), maxBytes)
	}
	setOpsUpstreamError(c, resp.StatusCode, upstreamMsg, upstreamDetail)

	if s.cfg != nil && s.cfg.Gateway.LogUpstreamErrorBody {
		log.Printf(
			"OpenAI upstream error %d (account=%d platform=%s type=%s): %s",
			resp.StatusCode,
			account.ID,
			account.Platform,
			account.Type,
			truncateForLog(body, s.cfg.Gateway.LogUpstreamErrorBodyMaxBytes),
		)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if errObj, ok := payload["error"].(map[string]any); ok {
			if msg, ok := errObj["message"].(string); ok {
				errObj["message"] = sanitizeUpstreamErrorMessage(msg)
			}
			c.JSON(resp.StatusCode, payload)
			return nil, fmt.Errorf("upstream error: %d", resp.StatusCode)
		}
	}

	s.writeChatErrorResponse(c, resp.StatusCode, "upstream_error", defaultUpstreamMessage(resp.StatusCode))
	return nil, fmt.Errorf("upstream error: %d message=%s", resp.StatusCode, upstreamMsg)
}

func (s *OpenAIGatewayService) writeChatErrorResponse(c *gin.Context, status int, errType, message string) {
	if c == nil {
		return
	}
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func defaultUpstreamMessage(statusCode int) string {
	switch statusCode {
	case 401:
		return "Upstream authentication failed, please contact administrator"
	case 402:
		return "Upstream payment required: insufficient balance or billing issue"
	case 403:
		return "Upstream access forbidden, please contact administrator"
	case 429:
		return "Upstream rate limit exceeded, please retry later"
	default:
		return "Upstream request failed"
	}
}
