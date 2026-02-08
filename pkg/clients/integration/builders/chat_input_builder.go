// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package integration_client_builders

import (
	"strings"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/rapidaai/pkg/commons"
	gorm_types "github.com/rapidaai/pkg/models/gorm/types"
	"github.com/rapidaai/pkg/parsers"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type inputChatBuilder struct {
	logger         commons.Logger
	templateParser parsers.StringTemplateParser
}

func NewChatInputBuilder(logger commons.Logger) InputChatBuilder {
	return &inputChatBuilder{
		logger:         logger,
		templateParser: parsers.NewPongo2StringTemplateParser(logger),
	}
}

func (in *inputChatBuilder) Credential(i uint64, dp *structpb.Struct) *protos.Credential {
	return &protos.Credential{
		Id:    i,
		Value: dp,
	}
}

func (in *inputChatBuilder) Chat(
	credential *protos.Credential,
	modelOpts map[string]*anypb.Any,
	tools []*protos.FunctionDefinition,
	additionalData map[string]string,
	conversations ...*protos.Message,
) *protos.ChatRequest {
	request := &protos.ChatRequest{
		Credential:      credential,
		Conversations:   conversations,
		ModelParameters: modelOpts,
		AdditionalData:  additionalData,
	}
	for _, tl := range tools {
		if request.ToolDefinitions == nil {
			request.ToolDefinitions = make([]*protos.ToolDefinition, 0)
		}
		request.ToolDefinitions = append(request.ToolDefinitions, &protos.ToolDefinition{
			Type:               "function",
			FunctionDefinition: tl,
		})
	}
	return request
}

func (in *inputChatBuilder) WithinMessage(role, prompt string) *protos.Message {
	return &protos.Message{
		Role: role,
		Contents: []*protos.Content{{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(prompt),
		}},
	}
}

func (in *inputChatBuilder) Message(
	template []*gorm_types.PromptTemplate,
	arguments map[string]interface{}) []*protos.Message {
	msg := make([]*protos.Message, 0)
	for _, v := range template {
		content := in.templateParser.Parse(v.GetContent(), arguments)
		msg = append(msg, in.WithinMessage(v.GetRole(), content))
	}
	return msg
}

func (in *inputChatBuilder) Arguments(variables []*gorm_types.PromptVariable, arguments map[string]*anypb.Any) map[string]interface{} {
	args, err := utils.AnyMapToInterfaceMap(arguments)
	if err != nil {
	}
	existing := in.PromptArguments(variables)
	return utils.MergeMaps(existing, args)
}

func (in *inputChatBuilder) PromptArguments(
	variables []*gorm_types.PromptVariable,
) map[string]interface{} {
	existing := make(map[string]interface{}, 0)
	for _, v := range variables {
		// Skip variables with dots in their name (e.g., "msg.role", "msg.message").
		// These are loop variable attributes auto-detected by the UI from pongo2
		// for-loop bodies like {{ msg.role }}. Adding them as top-level context keys
		// shadows the actual loop variable attribute resolution in pongo2.
		if strings.Contains(v.Name, ".") {
			continue
		}
		existing[v.Name] = v.DefaultValue
	}
	return existing
}

func (in *inputChatBuilder) Options(
	opts map[string]interface{},
	options map[string]*anypb.Any) map[string]*anypb.Any {
	// If options is nil, initialize it
	if options == nil {
		options = make(map[string]*anypb.Any)
	}

	for key, value := range opts {
		structValue, err := structpb.NewValue(value)
		if err != nil {
			continue
		}
		anyValue, err := anypb.New(structValue)
		if err != nil {
			continue
		}

		options[key] = anyValue
	}

	return options
}
