import { Metadata } from '@rapidaai/react';
import { SetMetadata } from '@/utils/metadata';

// ============================================================================
// Constants
// ============================================================================

const REQUIRED_KEYS = ['tool.endpoint_id', 'tool.parameters'];

// ============================================================================
// Default Options
// ============================================================================

export const GetEndpointDefaultOptions = (current: Metadata[]): Metadata[] => {
  const metadata: Metadata[] = [];

  const addMetadata = (key: string, defaultValue?: string) => {
    const meta = SetMetadata(current, key, defaultValue);
    if (meta) metadata.push(meta);
  };

  addMetadata('tool.endpoint_id');
  addMetadata('tool.parameters');

  return metadata.filter(m => REQUIRED_KEYS.includes(m.getKey()));
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
  const foundKeys = new Set(options.map(opt => opt.getKey()));
  const missingKeys = REQUIRED_KEYS.filter(key => !foundKeys.has(key));

  if (missingKeys.length > 0) {
    return `Please ensure all required metadata keys are present: ${REQUIRED_KEYS.join(', ')}.`;
  }
  return undefined;
};

const validateEndpointId = (value: string | undefined): string | undefined => {
  if (typeof value !== 'string' || value === '') {
    return 'Please provide a valid value for tool.endpoint_id. It must be a non-empty string.';
  }
  return undefined;
};

const validateParameters = (value: string | undefined): string | undefined => {
  if (typeof value !== 'string' || value === '') {
    return 'Please provide a valid value for tool.parameters. It must be a non-empty JSON string.';
  }

  try {
    const parameters = JSON.parse(value);

    if (
      typeof parameters !== 'object' ||
      parameters === null ||
      Array.isArray(parameters)
    ) {
      return 'Please ensure tool.parameters is a valid JSON object.';
    }

    const entries = Object.entries(parameters);
    if (entries.length === 0) {
      return 'Please provide parameter values within tool.parameters. It cannot be an empty object.';
    }

    for (const [paramKey, paramValue] of entries) {
      const [type, key] = paramKey.split('.');
      if (
        !type ||
        !key ||
        typeof paramValue !== 'string' ||
        paramValue === ''
      ) {
        return 'Please ensure each parameter key follows the format "type.key" and the values are non-empty strings.';
      }
    }

    // Check for unique values
    const values = entries.map(([, v]) => v);
    if (new Set(values).size !== values.length) {
      return 'Please ensure parameter values within tool.parameters are unique.';
    }
  } catch {
    return 'Please provide a valid JSON string for tool.parameters.';
  }

  return undefined;
};

export const ValidateEndpointDefaultOptions = (
  options: Metadata[],
): string | undefined => {
  // Run all validations in sequence, return first error
  return (
    validateRequiredKeys(options) ||
    validateEndpointId(getOptionValue(options, 'tool.endpoint_id')) ||
    validateParameters(getOptionValue(options, 'tool.parameters'))
  );
};
