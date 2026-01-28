import { Metadata } from '@rapidaai/react';
import { SetMetadata } from '@/utils/metadata';

export const GROQ_TEXT_MODEL = [
  {
    id: 'groq/llama-3.3-70b-versatile',
    name: 'llama-3.3-70b-versatile',
    human_name: 'Llama 3.3 70B Versatile',
    description: 'Most capable Llama model for complex tasks',
    category: 'text',
    status: 'ACTIVE',
  },
  {
    id: 'groq/llama-3.1-8b-instant',
    name: 'llama-3.1-8b-instant',
    human_name: 'Llama 3.1 8B Instant',
    description: 'Fastest model for real-time applications',
    category: 'text',
    status: 'ACTIVE',
  },
  {
    id: 'groq/llama-3.2-90b-vision-preview',
    name: 'llama-3.2-90b-vision-preview',
    human_name: 'Llama 3.2 90B Vision',
    description: 'Vision-capable Llama model for multimodal tasks',
    category: 'text',
    status: 'ACTIVE',
  },
  {
    id: 'groq/llama-3.2-11b-vision-preview',
    name: 'llama-3.2-11b-vision-preview',
    human_name: 'Llama 3.2 11B Vision',
    description: 'Compact vision-capable Llama model',
    category: 'text',
    status: 'ACTIVE',
  },
  {
    id: 'groq/mixtral-8x7b-32768',
    name: 'mixtral-8x7b-32768',
    human_name: 'Mixtral 8x7B (32K)',
    description: 'Mixture of experts with 32K context window',
    category: 'text',
    status: 'ACTIVE',
  },
  {
    id: 'groq/gemma2-9b-it',
    name: 'gemma2-9b-it',
    human_name: 'Gemma 2 9B',
    description: 'Google Gemma 2 instruction-tuned model',
    category: 'text',
    status: 'ACTIVE',
  },
  {
    id: 'groq/llama-guard-3-8b',
    name: 'llama-guard-3-8b',
    human_name: 'Llama Guard 3 8B',
    description: 'Safety and moderation model',
    category: 'text',
    status: 'ACTIVE',
  },
];

export const GetGroqTextProviderDefaultOptions = (
  current: Metadata[],
): Metadata[] => {
  const mtds: Metadata[] = [];
  const keysToKeep = [
    'rapida.credential_id',
    'model.id',
    'model.name',
    'model.frequency_penalty',
    'model.temperature',
    'model.top_p',
    'model.presence_penalty',
    'model.max_completion_tokens',
    'model.stop',
    'model.tool_choice',
    'model.response_format',
  ];

  const addMetadata = (
    key: string,
    defaultValue?: string,
    validationFn?: (value: string) => boolean,
  ) => {
    const metadata = SetMetadata(current, key, defaultValue, validationFn);
    if (metadata) mtds.push(metadata);
  };

  addMetadata('model.id', GROQ_TEXT_MODEL[0].id, value =>
    GROQ_TEXT_MODEL.some(model => model.id === value),
  );

  addMetadata('model.name', GROQ_TEXT_MODEL[0].name, value =>
    GROQ_TEXT_MODEL.some(model => model.name === value),
  );
  addMetadata('model.frequency_penalty', '0');
  addMetadata('model.temperature', '0.7');
  addMetadata('model.top_p', '1');
  addMetadata('model.presence_penalty', '0');
  addMetadata('model.max_completion_tokens', '2048');
  addMetadata('model.stop');
  addMetadata('model.tool_choice');
  addMetadata('model.response_format');
  addMetadata('rapida.credential_id');

  return mtds.filter(m => keysToKeep.includes(m.getKey()));
};

export const ValidateGroqTextProviderDefaultOptions = (
  options: Metadata[],
): string | undefined => {
  const credentialID = options.find(
    opt => opt.getKey() === 'rapida.credential_id',
  );
  if (
    !credentialID ||
    !credentialID.getValue() ||
    credentialID.getValue().length === 0
  ) {
    return 'Please check and provide valid credentials for Groq';
  }

  const modelIdOption = options.find(opt => opt.getKey() === 'model.id');
  if (
    !modelIdOption ||
    !GROQ_TEXT_MODEL.some(model => model.id === modelIdOption.getValue())
  ) {
    return 'Please check and select a valid model from dropdown.';
  }

  const frequencyPenaltyOption = options.find(
    opt => opt.getKey() === 'model.frequency_penalty',
  );
  if (
    !frequencyPenaltyOption ||
    isNaN(parseFloat(frequencyPenaltyOption.getValue())) ||
    parseFloat(frequencyPenaltyOption.getValue()) < -2 ||
    parseFloat(frequencyPenaltyOption.getValue()) > 2
  ) {
    return 'Please check and provide a correct value for frequency_penalty (between -2 and 2).';
  }

  const temperatureOption = options.find(
    opt => opt.getKey() === 'model.temperature',
  );
  if (
    !temperatureOption ||
    isNaN(parseFloat(temperatureOption.getValue())) ||
    parseFloat(temperatureOption.getValue()) < 0 ||
    parseFloat(temperatureOption.getValue()) > 2
  ) {
    return 'Please check and provide a correct value for temperature (between 0 and 2).';
  }

  const topPOption = options.find(opt => opt.getKey() === 'model.top_p');
  if (
    !topPOption ||
    isNaN(parseFloat(topPOption.getValue())) ||
    parseFloat(topPOption.getValue()) < 0 ||
    parseFloat(topPOption.getValue()) > 1
  ) {
    return 'Please check and provide a correct value for top_p (between 0 and 1).';
  }

  const presencePenaltyOption = options.find(
    opt => opt.getKey() === 'model.presence_penalty',
  );
  if (
    !presencePenaltyOption ||
    isNaN(parseFloat(presencePenaltyOption.getValue())) ||
    parseFloat(presencePenaltyOption.getValue()) < -2 ||
    parseFloat(presencePenaltyOption.getValue()) > 2
  ) {
    return 'Please check and provide a correct value for presence_penalty (between -2 and 2).';
  }

  const maxCompletionTokensOption = options.find(
    opt => opt.getKey() === 'model.max_completion_tokens',
  );
  if (
    !maxCompletionTokensOption ||
    isNaN(parseInt(maxCompletionTokensOption.getValue())) ||
    parseInt(maxCompletionTokensOption.getValue()) < 1
  ) {
    return 'Please check and provide a correct value for max_completion_tokens (should be greater than 1).';
  }

  const responseFormatOption = options.find(
    opt => opt.getKey() === 'model.response_format',
  );
  if (responseFormatOption && responseFormatOption.getValue()) {
    try {
      const parsedFormat = JSON.parse(responseFormatOption.getValue());
      if (typeof parsedFormat !== 'object' || !parsedFormat.type) {
        return 'Please check and provide a correct value for response_format (should be a valid JSON object).';
      }
      if (!['text', 'json_object'].includes(parsedFormat.type)) {
        return 'Please check and provide a correct value for response_format type (text or json_object).';
      }
    } catch {
      return 'Please check and provide a correct value for response_format.';
    }
  }

  return undefined;
};
