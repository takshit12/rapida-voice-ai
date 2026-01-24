import { Metadata } from '@rapidaai/react';
import { SetMetadata } from '@/utils/metadata';

// ============================================================================
// Constants
// ============================================================================

const REQUIRED_KEYS = ['mcp.server_url'];
const OPTIONAL_KEYS = [
  'mcp.tool_name',
  'mcp.protocol',
  'mcp.timeout',
  'mcp.headers',
];
const ALL_KEYS = [...REQUIRED_KEYS, ...OPTIONAL_KEYS];

export const MCP_PROTOCOL_OPTIONS = [
  { value: 'sse', name: 'SSE (Server-Sent Events)' },
  { value: 'streamable_http', name: 'Streamable HTTP' },
  { value: 'websocket', name: 'WebSocket' },
];

// ============================================================================
// Default Options
// ============================================================================

export const GetMCPDefaultOptions = (current: Metadata[]): Metadata[] => {
  const metadata: Metadata[] = [];

  const addMetadata = (key: string, defaultValue?: string) => {
    const meta = SetMetadata(current, key, defaultValue);
    if (meta) metadata.push(meta);
  };

  addMetadata('mcp.server_url');
  addMetadata('mcp.tool_name');
  addMetadata('mcp.protocol', 'sse');
  addMetadata('mcp.timeout', '30');
  addMetadata('mcp.headers');

  return metadata.filter(m => ALL_KEYS.includes(m.getKey()));
};

// ============================================================================
// Validation
// ============================================================================

const getOptionValue = (
  options: Metadata[],
  key: string,
): string | undefined => {
  return options.find(opt => opt.getKey() === key)?.getValue();
};

const validateRequiredKeys = (options: Metadata[]): string | undefined => {
  const missingKeys = REQUIRED_KEYS.filter(
    key => !options.some(opt => opt.getKey() === key),
  );

  if (missingKeys.length > 0) {
    return `Missing required configuration: ${missingKeys.join(', ')}`;
  }

  return undefined;
};

const validateServerUrl = (options: Metadata[]): string | undefined => {
  const serverUrl = getOptionValue(options, 'mcp.server_url');

  if (!serverUrl || serverUrl.trim() === '') {
    return 'MCP Server URL is required';
  }

  // Basic URL validation
  try {
    new URL(serverUrl);
  } catch {
    return 'Invalid MCP Server URL format';
  }

  // Ensure it's HTTP or HTTPS
  if (
    !serverUrl.startsWith('http://') &&
    !serverUrl.startsWith('https://') &&
    !serverUrl.startsWith('wss://')
  ) {
    return 'MCP Server URL must start with http://, https://, or wss://';
  }

  return undefined;
};

const validateProtocol = (options: Metadata[]): string | undefined => {
  const protocol = getOptionValue(options, 'mcp.protocol');

  if (
    protocol &&
    !['sse', 'websocket', 'streamable_http', ''].includes(protocol)
  ) {
    return 'Protocol must be either "sse", "websocket", or "streamable_http"';
  }

  return undefined;
};

const validateTimeout = (options: Metadata[]): string | undefined => {
  const timeout = getOptionValue(options, 'mcp.timeout');

  if (timeout) {
    const timeoutNum = parseInt(timeout, 10);
    if (isNaN(timeoutNum) || timeoutNum < 1 || timeoutNum > 300) {
      return 'Timeout must be a number between 1 and 300 seconds';
    }
  }

  return undefined;
};

const validateHeaders = (options: Metadata[]): string | undefined => {
  const headers = getOptionValue(options, 'mcp.headers');

  if (headers && headers.trim() !== '') {
    try {
      const parsed = JSON.parse(headers);
      if (typeof parsed !== 'object' || Array.isArray(parsed)) {
        return 'Headers must be a JSON object';
      }
    } catch {
      return 'Invalid JSON format for headers';
    }
  }

  return undefined;
};

export const ValidateMCPDefaultOptions = (
  options: Metadata[],
): string | undefined => {
  // Run all validations
  const validators = [
    validateRequiredKeys,
    validateServerUrl,
    validateProtocol,
    validateTimeout,
    validateHeaders,
  ];

  for (const validator of validators) {
    const error = validator(options);
    if (error) return error;
  }

  return undefined;
};
