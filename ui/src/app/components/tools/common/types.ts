import { Metadata } from '@rapidaai/react';

// ============================================================================
// Core Types
// ============================================================================

export interface ToolDefinition {
  name: string;
  description: string;
  parameters: string;
}

export interface ConfigureToolProps {
  toolDefinition?: ToolDefinition;
  onChangeToolDefinition?: (value: ToolDefinition) => void;
  parameters: Metadata[] | null;
  inputClass?: string;
  onParameterChange: (params: Metadata[]) => void;
}

// ============================================================================
// Parameter Types
// ============================================================================

export type ParameterType =
  | 'tool'
  | 'assistant'
  | 'conversation'
  | 'argument'
  | 'metadata'
  | 'option'
  | 'custom';

export interface KeyValueParameter {
  key: string;
  value: string;
}

// ============================================================================
// Constants
// ============================================================================

export const PARAMETER_TYPE_OPTIONS: Array<{
  name: string;
  value: ParameterType;
}> = [
  { name: 'Tool', value: 'tool' },
  { name: 'Assistant', value: 'assistant' },
  { name: 'Conversation', value: 'conversation' },
  { name: 'Argument', value: 'argument' },
  { name: 'Metadata', value: 'metadata' },
  { name: 'Option', value: 'option' },
  { name: 'Custom', value: 'custom' },
];

export const HTTP_METHOD_OPTIONS = [
  { name: 'GET', value: 'GET' },
  { name: 'POST', value: 'POST' },
  { name: 'PUT', value: 'PUT' },
  { name: 'PATCH', value: 'PATCH' },
] as const;

export const ASSISTANT_KEY_OPTIONS = [
  { name: 'Name', value: 'name' },
  { name: 'Prompt', value: 'prompt' },
] as const;

export const CONVERSATION_KEY_OPTIONS = [
  { name: 'Messages', value: 'messages' },
] as const;

export const TOOL_KEY_OPTIONS = [
  { name: 'Argument', value: 'argument' },
  { name: 'Name', value: 'name' },
] as const;
