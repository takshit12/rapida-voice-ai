import { Metadata } from '@rapidaai/react';
import { SetMetadata } from '@/utils/metadata';

// ============================================================================
// Constants
// ============================================================================

const REQUIRED_KEYS = ['mcp.server_url'];
const ALL_KEYS = [...REQUIRED_KEYS];

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
  if (!serverUrl.startsWith('http://') && !serverUrl.startsWith('https://')) {
    return 'MCP Server URL must start with http:// or https://';
  }

  return undefined;
};

export const ValidateMCPDefaultOptions = (
  options: Metadata[],
): string | undefined => {
  // Run all validations
  const validators = [validateRequiredKeys, validateServerUrl];

  for (const validator of validators) {
    const error = validator(options);
    if (error) return error;
  }

  return undefined;
};
