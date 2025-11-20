import { Metadata } from '@rapidaai/react';
import { SetMetadata } from '@/utils/metadata';
import { GEMINI_MODEL } from '@/providers';

export const GetGeminiTextProviderDefaultOptions = (
  current: Metadata[],
): Metadata[] => {
  const mtds: Metadata[] = [];
  const keysToKeep = [
    'rapida.credential_id',
    'model.id',
    'model.name',
    'model.frequency_penalty',
    'model.presence_penalty',

    'model.temperature',
    'model.top_p',
    'model.top_k',

    'model.candidate_count',
    'model.seed',

    'model.max_output_tokens',
    'model.response_format',
    'model.stop_sequences',
    //
  ];

  const addMetadata = (
    key: string,
    defaultValue?: string,
    validationFn?: (value: string) => boolean,
  ) => {
    const metadata = SetMetadata(current, key, defaultValue, validationFn);
    if (metadata) mtds.push(metadata);
  };
  addMetadata('model.id');
  addMetadata('model.name');

  addMetadata('model.candidate_count', '1');
  addMetadata('model.temperature', '1');
  addMetadata('model.max_output_tokens', '2048');
  addMetadata('model.stop_sequences', 'STOP!');
  addMetadata('model.presence_penalty', '0.0');
  addMetadata('model.frequency_penalty', '0.0');
  addMetadata('model.response_format');
  addMetadata('model.top_p');
  addMetadata('model.top_k');
  addMetadata('rapida.credential_id');
  addMetadata('model.seed');

  return mtds.filter(m => keysToKeep.includes(m.getKey()));
};
export const ValidateGeminiTextProviderDefaultOptions = (
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
    return 'Please check and provide a valid credentials for google.';
  }
  const modelIdOption = options.find(opt => opt.getKey() === 'model.id');
  if (
    !modelIdOption ||
    !GEMINI_MODEL().some(model => model.id === modelIdOption.getValue())
  ) {
    return 'Please check and select valid model from dropdown.';
  }

  const presence_penalty = options.find(
    opt => opt.getKey() === 'model.presence_penalty',
  );
  if (presence_penalty)
    if (
      isNaN(parseFloat(presence_penalty.getValue())) ||
      parseFloat(presence_penalty.getValue()) < -2 ||
      parseFloat(presence_penalty.getValue()) > 2
    ) {
      return 'Please check and provide a correct value for presence_penalty a valid value between -2 to 2.';
    }

  const frequencyPenaltyOption = options.find(
    opt => opt.getKey() === 'model.frequency_penalty',
  );
  if (frequencyPenaltyOption)
    if (
      isNaN(parseFloat(frequencyPenaltyOption.getValue())) ||
      parseFloat(frequencyPenaltyOption.getValue()) < -2 ||
      parseFloat(frequencyPenaltyOption.getValue()) > 2
    ) {
      return 'Please check and provide a correct value for frequency_penalty a valid value between -2 to 2.';
    }

  const temperatureOption = options.find(
    opt => opt.getKey() === 'model.temperature',
  );
  if (
    !temperatureOption ||
    isNaN(parseFloat(temperatureOption.getValue())) ||
    parseFloat(temperatureOption.getValue()) < 0 ||
    parseFloat(temperatureOption.getValue()) > 1
  ) {
    return 'Please check and provide a correct value for temperature any decimal value between 0 to 1';
  }

  const topPOption = options.find(opt => opt.getKey() === 'model.top_p');
  if (topPOption)
    if (
      isNaN(parseFloat(topPOption.getValue())) ||
      parseFloat(topPOption.getValue()) < 0 ||
      parseFloat(topPOption.getValue()) > 1
    ) {
      return 'Please check and provide a correct value for top_p any decimal value between 0 to 1';
    }

  const maxCompletionTokensOption = options.find(
    opt => opt.getKey() === 'model.max_output_tokens',
  );
  if (
    !maxCompletionTokensOption ||
    isNaN(parseInt(maxCompletionTokensOption.getValue())) ||
    parseInt(maxCompletionTokensOption.getValue()) < 1
  ) {
    return 'Please check and provide a correct value for max_completion_tokens';
  }

  const responseFormatOption = options.find(
    opt => opt.getKey() === 'model.response_format',
  );
  if (responseFormatOption) {
    try {
      const parsedFormat = JSON.parse(responseFormatOption.getValue());
      if (typeof parsedFormat !== 'object' || !parsedFormat.type) {
        return 'Please check and provide a correct value for response_format it should be a valid json object.';
      }
      if (!['text', 'json_object', 'json_schema'].includes(parsedFormat.type)) {
        return 'Please check and provide a correct value for response_format it should have type with text, json_object, json_schema.';
      }
      if (parsedFormat.type === 'json_schema' && !parsedFormat.json_schema) {
        return 'Please check and provide a correct value for response_format it should have valid json_schema.';
      }
    } catch (error) {
      return 'Please check and provide a correct value for response_format.';
    }
  }

  return undefined;
};
