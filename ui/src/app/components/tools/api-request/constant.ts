import { Metadata } from '@rapidaai/react';
import { SetMetadata } from '@/utils/metadata';

// ============================================================================
// Constants
// ============================================================================

const REQUIRED_KEYS = ['tool.method', 'tool.endpoint', 'tool.parameters'];
const ALL_KEYS = [...REQUIRED_KEYS, 'tool.headers'];
const VALID_HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'];

// ============================================================================
// Default Options
// ============================================================================

export const GetAPIRequestDefaultOptions = (
  current: Metadata[],
): Metadata[] => {
  const metadata: Metadata[] = [];

  const addMetadata = (key: string, defaultValue?: string) => {
    const meta = SetMetadata(current, key, defaultValue);
    if (meta) metadata.push(meta);
  };

  addMetadata('tool.method', 'POST');
  addMetadata('tool.endpoint');
  addMetadata('tool.headers');
  addMetadata('tool.parameters');

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
    return `Please provide all required metadata keys: ${REQUIRED_KEYS.join(', ')}.`;
  }
  return undefined;
};

const validateHttpMethod = (method: string | undefined): string | undefined => {
  if (method && !VALID_HTTP_METHODS.includes(method.toUpperCase())) {
    return `Please provide HTTP method provided. Supported methods are: ${VALID_HTTP_METHODS.join(', ')}.`;
  }
  return undefined;
};

const validateEndpoint = (endpoint: string | undefined): string | undefined => {
  if (endpoint) {
    try {
      new URL(endpoint);
    } catch {
      return 'Please provide a valid URL for the endpoint.';
    }
  }
  return undefined;
};

const validateHeaders = (headers: string | undefined): string | undefined => {
  if (!headers) return undefined;

  if (typeof headers !== 'string') {
    return 'Please provide valid headers as a string for creating the API request tool.';
  }

  try {
    const parsed = JSON.parse(headers);
    for (const [key, value] of Object.entries(parsed)) {
      if (
        typeof key !== 'string' ||
        typeof value !== 'string' ||
        key.trim() === '' ||
        (value as string).trim() === ''
      ) {
        return `Please provide a valid header entry detected. Header key and value must be non-empty strings. Key: ${key}, Value: ${value}.`;
      }
    }
  } catch {
    return 'Please provide valid headers.';
  }
  return undefined;
};

const validateParameters = (params: string | undefined): string | undefined => {
  if (typeof params !== 'string' || params === '') {
    return 'Please provide valid parameters as a non-empty string.';
  }

  try {
    const parsed = JSON.parse(params);

    if (
      typeof parsed !== 'object' ||
      parsed === null ||
      Array.isArray(parsed)
    ) {
      return 'Parameters must be a valid JSON object.';
    }

    const entries = Object.entries(parsed);
    if (entries.length === 0) {
      return 'Parameters object must contain at least one key-value pair.';
    }

    for (const [paramKey, paramValue] of entries) {
      const [type, key] = paramKey.split('.');
      if (
        !type ||
        !key ||
        typeof paramValue !== 'string' ||
        paramValue === ''
      ) {
        return `Please provide a valid parameter format. Key: ${paramKey}, Value: ${paramValue}. Ensure key is in "type.key" format and value is a non-empty string.`;
      }
    }

    // Check for unique values
    const values = entries.map(([, value]) => value);
    if (new Set(values).size !== values.length) {
      return 'Please provide a valid parameter, values must be unique.';
    }
  } catch {
    return 'Please provide valid parameters, must be a valid JSON object.';
  }

  return undefined;
};

export const ValidateAPIRequestDefaultOptions = (
  options: Metadata[],
): string | undefined => {
  // Run all validations in sequence, return first error
  return (
    validateRequiredKeys(options) ||
    validateHttpMethod(getOptionValue(options, 'tool.method')) ||
    validateEndpoint(getOptionValue(options, 'tool.endpoint')) ||
    validateHeaders(getOptionValue(options, 'tool.headers')) ||
    validateParameters(getOptionValue(options, 'tool.parameters'))
  );
};
