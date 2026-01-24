import { SetMetadata } from '@/utils/metadata';
import { Metadata } from '@rapidaai/react';
import { AZURE_SPEECH_TO_TEXT_LANGUAGE } from '@/providers';

export const GetAzureDefaultOptions = (current: Metadata[]): Metadata[] => {
  const mtds: Metadata[] = [];
  // Define the keys we want to keep
  const keysToKeep = ['rapida.credential_id', 'listen.language'];

  // Function to create or update metadata

  const addMetadata = (
    key: string,
    defaultValue?: string,
    validationFn?: (value: string) => boolean,
  ) => {
    const metadata = SetMetadata(current, key, defaultValue, validationFn);
    if (metadata) mtds.push(metadata);
  };
  addMetadata('rapida.credential_id');
  // Set language
  addMetadata('listen.language', 'en', value =>
    AZURE_SPEECH_TO_TEXT_LANGUAGE().some(l => l.code === value),
  );

  // Only return metadata for the keys we want to keep
  return [
    ...mtds.filter(m => keysToKeep.includes(m.getKey())),
    ...current.filter(m => m.getKey().startsWith('microphone.')),
  ];
};

export const ValidateAzureOptions = (
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
    return 'Please provide a valid azure credential for speech to text.';
  }
  // Validate language
  const languageOption = options.find(
    opt => opt.getKey() === 'listen.language',
  );
  if (
    !languageOption ||
    !AZURE_SPEECH_TO_TEXT_LANGUAGE().some(
      lang => lang.code === languageOption.getValue(),
    )
  ) {
    return 'Please provide a valid azure language for speech to text.';
  }

  return undefined;
};
